package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/offers"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// ProvidersJob is the job_runs.job of the watch-provider stage (RFC-003 section 3). Its source is tmdb, so it runs
// only while both FLAG_INGEST_ENABLED and FLAG_INGEST_TMDB are on.
const ProvidersJob = "offers.tmdb"

// ProvidersBatch is how many items one Step call asks TMDB about.
const ProvidersBatch = 100

const maxProviderName = 80

// watchHost is the only host a stored link may point to: TMDB's watch page, which names JustWatch as the source.
const watchHost = "www.themoviedb.org"

type providerEntry struct {
	ID   int    `json:"provider_id"`
	Name string `json:"provider_name"`
}

type countryEntry struct {
	Link     string          `json:"link"`
	Flatrate []providerEntry `json:"flatrate"`
	Rent     []providerEntry `json:"rent"`
	Buy      []providerEntry `json:"buy"`
	Ads      []providerEntry `json:"ads"`
	Free     []providerEntry `json:"free"`
}

// ParseProviders reads a /watch/providers response for the given countries. TMDB's flatrate is a stream, ads and
// free are both free. A country whose link is not an https TMDB watch page is skipped as a whole, and an entry without
// a numeric id or a name is dropped, so nothing odd in the answer reaches the redirect or the page.
func ParseProviders(raw []byte, countries []string) ([]offers.Provider, error) {
	var doc struct {
		Results map[string]countryEntry `json:"results"`
	}

	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("decode tmdb watch providers: %w", err)
	}

	var out []offers.Provider

	for _, country := range countries {
		c, ok := doc.Results[country]
		if !ok || !validLink(c.Link) {
			continue
		}

		for _, group := range []struct {
			kind    string
			entries []providerEntry
		}{
			{"stream", c.Flatrate}, {"rent", c.Rent}, {"buy", c.Buy}, {"free", slices.Concat(c.Ads, c.Free)},
		} {
			for _, e := range group.entries {
				name := strings.TrimSpace(e.Name)
				if e.ID < 1 || name == "" {
					continue
				}

				out = append(out, offers.Provider{
					Country: country, Kind: group.kind, ID: e.ID, Name: truncate(name, maxProviderName), Link: c.Link,
				})
			}
		}
	}

	return out, nil
}

func validLink(raw string) bool {
	u, err := url.Parse(raw)

	return err == nil && u.Scheme == "https" && u.Host == watchHost && u.User == nil
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}

	return string(r[:n])
}

// dueSQL: $1 cursor (item id, empty = start), $2 asked before this time or never, $3 limit. Only public films and
// series with a TMDB id the catalogue already trusts (a number).
const dueSQL = `SELECT i.id::text, i.type, i.external_ids ->> 'tmdb'
FROM items i
LEFT JOIN offer_source_checks c ON c.item_id = i.id AND c.source = 'tmdb_providers'
WHERE i.type IN ('movie', 'series') AND i.visibility = 'public' AND (i.external_ids ->> 'tmdb') ~ '^[0-9]{1,9}$'
  AND ($1 = '' OR i.id > $1::uuid) AND (c.checked_at IS NULL OR c.checked_at < $2)
ORDER BY i.id LIMIT $3`

type providerTarget struct{ id, itemType, tmdbID string }

type providersCursor struct {
	After string `json:"after"`
}

// ProvidersStep asks TMDB for the watch providers of the films and series that were not asked about in the last
// ProviderRefresh, ProvidersBatch per call, in item id order (shortcut: not newest releases first, which a keyset
// cursor cannot give; upgrade if availability of new releases lags). One request per item. A rejected token stops the
// run; any other failure of an item goes to the dead letter and the item is asked about again on the next run.
func ProvidersStep(c *Connector, client *ingest.Client, now func() time.Time) run.Step {
	return func(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
		var cur providersCursor
		if len(r.Cursor) > 0 && string(r.Cursor) != "null" {
			if err := json.Unmarshal(r.Cursor, &cur); err != nil {
				return run.Result{}, fmt.Errorf("offers tmdb: cursor: %w: %w", err, queue.ErrPermanent)
			}
		}

		t := now().UTC()

		targets, err := dueProviders(ctx, tx, cur.After, t.Add(-offers.ProviderRefresh))
		if err != nil {
			return run.Result{}, err
		}

		res := run.Result{Done: len(targets) < ProvidersBatch, Targets: len(targets), Seen: len(targets)}

		for _, tg := range targets {
			cur.After = tg.id

			if err := ctx.Err(); err != nil {
				return run.Result{}, err
			}

			path := "/" + mediaOf(tg.itemType) + "/" + tg.tmdbID + "/watch/providers"

			body, found, err := c.get(ctx, client, path, url.Values{})
			if errors.Is(err, queue.ErrPermanent) {
				return run.Result{}, err
			}

			if err != nil {
				res.Failed++

				if _, ferr := run.RecordFailure(ctx, tx, r.ID, ProvidersJob, tg.id, "fetch_failed", err); ferr != nil {
					return run.Result{}, ferr
				}

				continue
			}

			var ps []offers.Provider

			if found {
				if ps, err = ParseProviders(body, offers.Countries); err != nil {
					res.Failed++

					if _, ferr := run.RecordFailure(
						ctx,
						tx,
						r.ID,
						ProvidersJob,
						tg.id,
						"invalid_payload",
						err,
					); ferr != nil {
						return run.Result{}, ferr
					}

					continue
				}
			}

			if err := offers.SaveProviders(ctx, tx, tg.id, t, ps); err != nil {
				return run.Result{}, err
			}

			// Changed counts items, like Seen: the digest compares the two (it used to add every provider offer).
			if len(ps) > 0 {
				res.Changed++
			}
		}

		res.Cursor, err = json.Marshal(cur)

		return res, err
	}
}

func dueProviders(ctx context.Context, tx *database.Tx, after string, askedBefore time.Time) ([]providerTarget, error) {
	rows, err := tx.QueryContext(ctx, dueSQL, after, askedBefore, ProvidersBatch)
	if err != nil {
		return nil, fmt.Errorf("offers tmdb: list due: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []providerTarget

	for rows.Next() {
		var tg providerTarget
		if err := rows.Scan(&tg.id, &tg.itemType, &tg.tmdbID); err != nil {
			return nil, err
		}

		if _, err := strconv.Atoi(tg.tmdbID); err != nil {
			continue
		}

		out = append(out, tg)
	}

	return out, rows.Err()
}
