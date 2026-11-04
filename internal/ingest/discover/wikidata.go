// Package discover finds catalogue items the top-N Wikidata selection never reaches. It walks Wikidata (CC0, free, no
// key) downwards by popularity (sitelink count), one type at a time, and stores the records of the items it has not
// fetched yet; the entities, the normaliser and the covers chain then treat them like any other item. It uses no
// other site: nothing here scrapes a source that forbids it (Goodreads, Skoob, JustWatch) or reads datasets with a
// commercial fee (IMDb), and it needs no TMDB or Open Library access (docs/data/SOURCES.md, ADR 0006).
package discover

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
)

// Job is the job_runs.job of the step; its source is wikidata, so the ingest_wikidata kill switch gates it.
const Job = "discover.wikidata"

// StateKey is the ingest_host_state row that keeps the position across runs.
const StateKey = "discover:wikidata"

const (
	// BatchSize is how many Wikidata items one chunk selects.
	BatchSize = 500
	// MinSitelinks is the popularity floor: an item on fewer than this many Wikipedia and sister-project pages is
	// mostly a stub, and the walk of its type ends there.
	MinSitelinks = 3
	// unbounded starts a type whose records are not in the database yet.
	unbounded = 1 << 30
)

// Fetcher resolves Wikidata QIDs into records (wikidata.Connector).
type Fetcher interface {
	Fetch(ctx context.Context, c *ingest.Client, batch []ingest.Target) ([]ingest.Record, []ingest.TargetError)
}

// Config wires the step. Zero values take the defaults; SparqlURL is overridden in tests.
type Config struct {
	Client       *ingest.Client
	Fetcher      Fetcher
	SparqlURL    string
	Batch        int
	MinSitelinks int
}

// position is where the walk of one type stands: the sitelink count and the item IRI of the last selected item.
type position struct {
	Links int    `json:"links"`
	After string `json:"after"`
	Done  bool   `json:"done"`
}

type state struct {
	Turn  int                 `json:"turn"`
	Types map[string]position `json:"types"`
}

// Row is one selected item.
type Row struct {
	QID   string
	IRI   string
	Links int
}

// Query selects the next items of a type below the position, most popular first. Items at the same sitelink count
// are ordered by IRI, so the pair (links, iri) is a total order a cursor can follow.
func Query(itemType string, links int, after string, floor, limit int) string {
	classes := make([]string, len(wikidata.TypeClasses[itemType]))
	for i, c := range wikidata.TypeClasses[itemType] {
		classes[i] = "wd:" + c
	}

	return fmt.Sprintf(`SELECT DISTINCT ?item ?links WHERE {
  VALUES ?class { %s }
  ?item wdt:P31 ?class ; wikibase:sitelinks ?links .
  FILTER(?links >= %d && (?links < %d || (?links = %d && STR(?item) > %q)))
} ORDER BY DESC(?links) ?item LIMIT %d`, strings.Join(classes, " "), floor, links, links, after, limit)
}

type sparqlResult struct {
	Results struct {
		Bindings []struct {
			Item  struct{ Value string } `json:"item"`
			Links struct{ Value string } `json:"links"`
		} `json:"bindings"`
	} `json:"results"`
}

// parseRows reads the SPARQL answer into rows in the order received.
func parseRows(raw []byte) ([]Row, error) {
	var res sparqlResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("discover: sparql answer: %w", err)
	}

	rows := make([]Row, 0, len(res.Results.Bindings))

	for _, b := range res.Results.Bindings {
		links, err := strconv.Atoi(b.Links.Value)
		if err != nil {
			return nil, fmt.Errorf("discover: sitelinks %q: %w", b.Links.Value, err)
		}

		i := strings.LastIndex(b.Item.Value, "/")
		rows = append(rows, Row{QID: b.Item.Value[i+1:], IRI: b.Item.Value, Links: links})
	}

	return rows, nil
}

func (c *Config) query(ctx context.Context, q string) ([]Row, error) {
	endpoint := c.SparqlURL
	if endpoint == "" {
		endpoint = wikidata.SparqlURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		endpoint+"?"+url.Values{"query": {q}, "format": {"json"}}.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := c.Client.JSON.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discover sparql: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discover sparql: status %d", resp.StatusCode)
	}

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("discover sparql: %w", err)
	}

	return parseRows(raw)
}

// startLinks is the least popular item of the type already stored: the walk starts there, below the top-N selection.
func startLinks(ctx context.Context, tx *database.Tx, itemType string) (int, error) {
	var links *int

	err := tx.QueryRowContext(ctx, `SELECT min((payload->>'sitelinks_count')::int) FROM source_records
WHERE source = 'wikidata' AND payload->>'type' = $1`, itemType).Scan(&links)
	if err != nil {
		return 0, fmt.Errorf("discover: start of %s: %w", itemType, err)
	}

	if links == nil {
		return unbounded, nil
	}

	return *links, nil
}

// Step handles one chunk: the next BatchSize items of the next type still being walked, minus the ones already
// fetched. The position is saved in the same transaction, so an interrupted run replays at most one chunk.
func (c *Config) Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	st := state{Types: map[string]position{}}
	if _, err := ingest.LoadState(ctx, tx, StateKey, &st); err != nil {
		return run.Result{}, err
	}

	if st.Types == nil {
		st.Types = map[string]position{}
	}

	itemType := ""

	for i := range wikidata.Types {
		t := wikidata.Types[(st.Turn+i)%len(wikidata.Types)]
		if !st.Types[t].Done {
			itemType, st.Turn = t, (st.Turn+i+1)%len(wikidata.Types)

			break
		}
	}

	if itemType == "" {
		return run.Result{Done: true}, nil
	}

	pos, started := st.Types[itemType]
	if !started {
		links, err := startLinks(ctx, tx, itemType)
		if err != nil {
			return run.Result{}, err
		}

		pos = position{Links: links}
	}

	limit, floor := c.Batch, c.MinSitelinks
	if limit <= 0 {
		limit = BatchSize
	}

	if floor <= 0 {
		floor = MinSitelinks
	}

	rows, err := c.query(ctx, Query(itemType, pos.Links, pos.After, floor, limit))
	if err != nil {
		return run.Result{}, err
	}

	res := run.Result{Targets: len(rows)}

	if len(rows) > 0 {
		last := rows[len(rows)-1]
		pos.Links, pos.After = last.Links, last.IRI
	}

	// A short page means the band down to the floor is exhausted.
	pos.Done = len(rows) < limit
	st.Types[itemType] = pos

	qids := make([]string, len(rows))
	for i, row := range rows {
		qids[i] = row.QID
	}

	fresh, err := notFetched(ctx, tx, qids)
	if err != nil {
		return run.Result{}, err
	}

	if len(fresh) > 0 {
		targets := make([]ingest.Target, len(fresh))
		for i, q := range fresh {
			targets[i] = ingest.Target{ID: q, Type: itemType}
		}

		records, failures := c.Fetcher.Fetch(ctx, c.Client, targets)
		for _, rec := range records {
			changed, err := store.Upsert(ctx, tx, r.ID, rec)
			if err != nil {
				return run.Result{}, err
			}

			res.Seen++

			if changed {
				res.Changed++
			}
		}

		for _, f := range failures {
			res.Failed++

			if _, err := run.RecordFailure(ctx, tx, r.ID, Job, f.Target.ID, "fetch_failed", f.Err); err != nil {
				return run.Result{}, err
			}
		}
	}

	return res, ingest.SaveState(ctx, tx, StateKey, st)
}

// notFetched keeps the QIDs that have no Wikidata record yet.
func notFetched(ctx context.Context, tx *database.Tx, qids []string) ([]string, error) {
	if len(qids) == 0 {
		return nil, nil
	}

	rows, err := tx.QueryContext(ctx, `SELECT q FROM unnest($1::text[]) AS u(q)
WHERE NOT EXISTS (SELECT 1 FROM source_records s WHERE s.source = 'wikidata' AND s.external_id = u.q)`,
		"{"+strings.Join(qids, ",")+"}")
	if err != nil {
		return nil, fmt.Errorf("discover: known records: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []string

	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, err
		}

		out = append(out, q)
	}

	return out, rows.Err()
}
