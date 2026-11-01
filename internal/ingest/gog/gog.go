// Package gog is the GOG catalogue connector (SDD 06 §7.3): store link, vertical cover and OS platforms for games
// with a GOG id (Wikidata P2725). Wikidata stores the store path (game/<slug>), which the catalogue cannot filter
// on, so the connector searches by English title and matches by slug, then title, then slug prefix. Port of legacy
// connectors/gog.py.
//
// Terms: GOG publishes no API terms for the public catalogue endpoint; the GOG User Agreement
// (https://www.gog.com/en/user_agreement, not reachable from a plain fetch on 2026-11-01) governs the site. Owner
// cleared the read in Q-A16. Images stay site-only (SDD 06 §7.1); the GOG id comes from Wikidata.
package gog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// APIURL is the public catalogue search.
const APIURL = "https://catalog.gog.com/v1/catalog"

var (
	_ ingest.Connector = (*Connector)(nil)

	nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	osNames  = map[string]string{"windows": "Windows", "osx": "macOS", "linux": "Linux"}
)

// Payload is the stored record body.
type Payload struct {
	Qid       string   `json:"qid"`
	ID        string   `json:"id"`
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	StoreURL  string   `json:"store_url"`
	Cover     string   `json:"cover,omitempty"`
	Platforms []string `json:"platforms"`
	State     string   `json:"state,omitempty"`
}

type product struct {
	ID               string   `json:"id"`
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	StoreLink        string   `json:"storeLink"`
	CoverVertical    string   `json:"coverVertical"`
	OperatingSystems []string `json:"operatingSystems"`
	ProductState     string   `json:"productState"`
}

func norm(s string) string { return nonAlnum.ReplaceAllString(strings.ToLower(s), "") }

// ParseCatalog picks the product of gogID (game/<slug>) from a search response: exact slug, then normalised title,
// then slug prefix. nil without error when nothing matches.
func ParseCatalog(raw []byte, gogID, title, qid string) (*Payload, error) {
	var doc struct {
		Products []product `json:"products"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("decode gog catalog: %w", err)
	}

	slug := strings.TrimPrefix(strings.TrimPrefix(gogID, "game/"), "movie/")

	for _, match := range []func(product) bool{
		func(p product) bool { return p.Slug == slug },
		func(p product) bool { return norm(p.Title) == norm(title) },
		func(p product) bool { return strings.HasPrefix(p.Slug, slug) },
	} {
		for _, p := range doc.Products {
			if !match(p) {
				continue
			}

			out := &Payload{
				Qid: qid, ID: p.ID, Slug: p.Slug, Title: p.Title, StoreURL: p.StoreLink,
				Cover: p.CoverVertical, State: p.ProductState, Platforms: []string{},
			}

			for _, o := range p.OperatingSystems {
				if name, ok := osNames[o]; ok {
					out.Platforms = append(out.Platforms, name)
				}
			}

			return out, nil
		}
	}

	return nil, nil
}

// Connector searches the GOG catalogue for stored games. A Target carries the QID in ID and Ref(...) in Type.
type Connector struct {
	// APIURL overrides APIURL; tests only.
	APIURL string
}

// Source names the connector in source_records.
func (*Connector) Source() string { return "gog" }

// Ref packs the GOG id and the English title; GOG ids never contain "|".
func Ref(gogID, title string) string { return gogID + "|" + title }

// targetsSQL: $1 after, $2 since (null = any), $3 only missing, $4 limit (0 = none).
const targetsSQL = `
SELECT i.wikidata_qid, i.external_ids->>'gog', t.title
FROM items i JOIN item_translations t ON t.item_id = i.id AND t.locale = 'en'
WHERE i.type = 'game' AND i.external_ids ? 'gog'
  AND i.wikidata_qid > $1
  AND ($2::timestamptz IS NULL OR i.updated_at >= $2)
  AND (NOT $3::boolean OR NOT EXISTS (
    SELECT 1 FROM source_records s WHERE s.source = 'gog' AND s.external_id = i.wikidata_qid))
ORDER BY i.wikidata_qid
LIMIT CASE WHEN $4::int > 0 THEN $4::int END`

// Targets yields the games with a GOG id and an English title, after the cursor.
func (*Connector) Targets(
	ctx context.Context, db database.Executor, p ingest.RunParams, after string,
) iter.Seq2[ingest.Target, error] {
	return func(yield func(ingest.Target, error) bool) {
		var since any
		if p.Mode == run.ModeIncremental && !p.Since.IsZero() {
			since = p.Since
		}

		rows, err := db.QueryContext(ctx, targetsSQL, after, since, p.Mode == run.ModeBootstrap, p.Limit)
		if err != nil {
			yield(ingest.Target{}, fmt.Errorf("gog targets: %w", err))

			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var qid, id, title string
			if err := rows.Scan(&qid, &id, &title); err != nil {
				yield(ingest.Target{}, fmt.Errorf("scan gog target: %w", err))

				return
			}

			if !yield(ingest.Target{ID: qid, Type: Ref(id, title)}, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(ingest.Target{}, fmt.Errorf("gog targets: %w", err))
		}
	}
}

func (c *Connector) one(ctx context.Context, client *ingest.Client, t ingest.Target) (*ingest.Record, error) {
	gogID, title, _ := strings.Cut(t.Type, "|")

	endpoint := c.APIURL
	if endpoint == "" {
		endpoint = APIURL
	}

	q := url.Values{
		"limit": {"10"}, "query": {"like:" + title}, "productType": {"in:game,pack"},
		"countryCode": {"US"}, "locale": {"en-US"}, "currencyCode": {"USD"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build gog request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.JSON.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gog %s: %w", t.ID, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gog %s: status %d", t.ID, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gog %s: %w", t.ID, err)
	}

	p, err := ParseCatalog(body, gogID, title, t.ID)
	if err != nil || p == nil {
		return nil, err
	}

	raw, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("encode gog %s: %w", t.ID, err)
	}

	return &ingest.Record{Source: c.Source(), ExternalID: t.ID, Payload: raw}, nil
}

// Fetch resolves the batch one search at a time (2/s, SDD 06 §7.2). A game without a matching product is skipped;
// a failed request is kept for the dead-letter table.
func (c *Connector) Fetch(
	ctx context.Context, client *ingest.Client, batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	var (
		records []ingest.Record
		errs    []ingest.TargetError
	)

	for _, t := range batch {
		rec, err := c.one(ctx, client, t)

		switch {
		case err != nil:
			errs = append(errs, ingest.TargetError{Target: t, Err: err})
		case rec != nil:
			records = append(records, *rec)
		}
	}

	return records, errs
}
