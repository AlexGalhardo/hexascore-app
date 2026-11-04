// Package discover finds catalogue items the top-N Wikidata selection never reaches, by walking an official id list
// in order. It uses only sources whose terms allow it: the daily id exports TMDB publishes for exactly this purpose
// (https://developer.themoviedb.org/docs/daily-id-exports) and Wikidata (CC0). Nothing here scrapes a site that
// forbids it (Goodreads, Skoob, JustWatch) or reads the IMDb datasets (docs/data/SOURCES.md, ADR 0006).
//
// A TMDB id with no Wikidata item (property P4947 for films, P4983 for series) is skipped: items are keyed by their
// Wikidata QID, so the normaliser, the entities and the covers chain treat a discovered item like any other.
package discover

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// Job is the job_runs.job of the step; its source is tmdb, so the ingest_tmdb kill switch gates it.
const Job = "discover.tmdb"

// StateKey is the ingest_host_state row that keeps the position across runs.
const StateKey = "discover:tmdb"

const (
	// ExportBase is where TMDB publishes the daily id exports.
	ExportBase = "https://files.tmdb.org/p/exports"
	// BatchSize is how many TMDB ids one chunk resolves.
	BatchSize = 100
	// MinPopularity leaves the long tail of empty entries out: a TMDB entry below it is mostly a stub.
	MinPopularity = 1.0
	// maxExportAge is how many days back the export is looked for (it is published each morning, UTC).
	maxExportAge = 3
	// sparqlBatch bounds the VALUES list of one Wikidata query.
	sparqlBatch = 100
)

// kinds in processing order: the export name, the Wikidata property of the id, the HexaScore item type.
var kinds = []struct{ Name, Export, Prop, Type string }{
	{"movie", "movie_ids", "P4947", "movie"},
	{"tv", "tv_series_ids", "P4983", "series"},
}

// Fetcher resolves Wikidata QIDs into records (wikidata.Connector).
type Fetcher interface {
	Fetch(ctx context.Context, c *ingest.Client, batch []ingest.Target) ([]ingest.Record, []ingest.TargetError)
}

// Config wires the step. Zero values take the defaults; the URLs are overridden in tests.
type Config struct {
	Client        *ingest.Client
	Fetcher       Fetcher
	Now           func() time.Time
	ExportBase    string
	SparqlURL     string
	MinPopularity float64
	Batch         int

	mu    sync.Mutex
	cache map[string][]int
}

// state is the position kept between runs: the last TMDB id handled per kind, and the day each kind was exhausted.
type state struct {
	After  map[string]int    `json:"after"`
	DoneOn map[string]string `json:"done_on"`
}

type exportLine struct {
	ID         int     `json:"id"`
	Adult      bool    `json:"adult"`
	Popularity float64 `json:"popularity"`
}

// ParseExport reads a daily export (line-delimited JSON, already gunzipped) into ascending ids, leaving out adult
// entries and the ones below minPopularity.
func ParseExport(r io.Reader, minPopularity float64) ([]int, error) {
	var ids []int

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)

	for sc.Scan() {
		var l exportLine
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil {
			return nil, fmt.Errorf("tmdb export: %w", err)
		}

		if l.Adult || l.Popularity < minPopularity || l.ID <= 0 {
			continue
		}

		ids = append(ids, l.ID)
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("tmdb export: %w", err)
	}

	sort.Ints(ids)

	return slices.Compact(ids), nil
}

// NextBatch returns up to n ids after the given one in an ascending list.
func NextBatch(ids []int, after, n int) []int {
	i := sort.SearchInts(ids, after+1)
	if i >= len(ids) {
		return nil
	}

	return ids[i:min(i+n, len(ids))]
}

// exportURL is the file of the day: <base>/<movie_ids|tv_series_ids>_MM_DD_YYYY.json.gz.
func exportURL(base, name string, day time.Time) string {
	return fmt.Sprintf("%s/%s_%s.json.gz", base, name, day.Format("01_02_2006"))
}

func (c *Config) ids(ctx context.Context, kind int, now time.Time) ([]int, error) {
	key := kinds[kind].Name + now.UTC().Format(time.DateOnly)

	c.mu.Lock()
	defer c.mu.Unlock()

	if got, ok := c.cache[key]; ok {
		return got, nil
	}

	base := c.ExportBase
	if base == "" {
		base = ExportBase
	}

	var lastErr error

	for age := 1; age <= maxExportAge; age++ {
		u := exportURL(base, kinds[kind].Export, now.UTC().AddDate(0, 0, -age))

		ids, err := c.download(ctx, u)
		if err != nil {
			lastErr = err

			continue
		}

		c.cache = map[string][]int{key: ids} // one export in memory at a time

		return ids, nil
	}

	return nil, fmt.Errorf("discover: no %s export in the last %d days: %w", kinds[kind].Name, maxExportAge, lastErr)
}

// download fetches and parses one export file.
func (c *Config) download(ctx context.Context, u string) ([]int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.JSON.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	zr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tmdb export gzip: %w", err)
	}

	defer func() { _ = zr.Close() }()

	return ParseExport(zr, c.minPopularity())
}

func (c *Config) minPopularity() float64 {
	if c.MinPopularity > 0 {
		return c.MinPopularity
	}

	return MinPopularity
}

// sparql returns the query that maps TMDB ids to Wikidata items through prop.
func sparql(prop string, ids []int) string {
	vals := make([]string, len(ids))
	for i, id := range ids {
		vals[i] = `"` + strconv.Itoa(id) + `"`
	}

	return fmt.Sprintf(`SELECT ?item ?id WHERE { VALUES ?id { %s } ?item wdt:%s ?id }`, strings.Join(vals, " "), prop)
}

type sparqlResult struct {
	Results struct {
		Bindings []struct {
			Item struct {
				Value string `json:"value"`
			} `json:"item"`
		} `json:"bindings"`
	} `json:"results"`
}

// resolve maps TMDB ids to Wikidata QIDs; an id with no Wikidata item is absent from the result.
func (c *Config) resolve(ctx context.Context, prop string, ids []int) ([]string, error) {
	endpoint := c.SparqlURL
	if endpoint == "" {
		endpoint = "https://query.wikidata.org/sparql"
	}

	var qids []string

	for chunk := range slices.Chunk(ids, sparqlBatch) {
		res, err := c.query(ctx, endpoint, sparql(prop, chunk))
		if err != nil {
			return nil, err
		}

		for _, b := range res.Results.Bindings {
			if i := strings.LastIndex(b.Item.Value, "/"); i >= 0 {
				qids = append(qids, b.Item.Value[i+1:])
			}
		}
	}

	slices.Sort(qids)

	return slices.Compact(qids), nil
}

// query runs one SPARQL query against the endpoint.
func (c *Config) query(ctx context.Context, endpoint, q string) (sparqlResult, error) {
	var res sparqlResult

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		endpoint+"?"+url.Values{"query": {q}, "format": {"json"}}.Encode(), nil)
	if err != nil {
		return res, err
	}

	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := c.Client.JSON.Do(req)
	if err != nil {
		return res, fmt.Errorf("discover sparql: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("discover sparql: status %d", resp.StatusCode)
	}

	return res, json.NewDecoder(resp.Body).Decode(&res)
}

// Step handles one chunk: the next BatchSize TMDB ids of the first kind not yet exhausted today, resolved to
// Wikidata items not fetched before, whose records are stored for the normaliser. The position is saved in the same
// transaction, so an interrupted run replays at most one chunk.
func (c *Config) Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	now := c.Now().UTC()
	today := now.Format(time.DateOnly)

	st := state{After: map[string]int{}, DoneOn: map[string]string{}}
	if _, err := ingest.LoadState(ctx, tx, StateKey, &st); err != nil {
		return run.Result{}, err
	}

	if st.After == nil {
		st.After = map[string]int{}
	}

	if st.DoneOn == nil {
		st.DoneOn = map[string]string{}
	}

	kind := -1

	for i, k := range kinds {
		if st.DoneOn[k.Name] != today {
			kind = i

			break
		}
	}

	if kind < 0 {
		return run.Result{Done: true}, nil
	}

	k := kinds[kind]

	ids, err := c.ids(ctx, kind, now)
	if err != nil {
		return run.Result{}, err
	}

	n := c.Batch
	if n <= 0 {
		n = BatchSize
	}

	batch := NextBatch(ids, st.After[k.Name], n)
	res := run.Result{Targets: len(batch)}

	if len(batch) == 0 {
		st.DoneOn[k.Name] = today
		res.Done = kind == len(kinds)-1

		return res, ingest.SaveState(ctx, tx, StateKey, st)
	}

	qids, err := c.resolve(ctx, k.Prop, batch)
	if err != nil {
		return run.Result{}, err
	}

	fresh, err := notFetched(ctx, tx, qids)
	if err != nil {
		return run.Result{}, err
	}

	st.After[k.Name] = batch[len(batch)-1]

	if len(fresh) > 0 {
		targets := make([]ingest.Target, len(fresh))
		for i, q := range fresh {
			targets[i] = ingest.Target{ID: q, Type: k.Type}
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
