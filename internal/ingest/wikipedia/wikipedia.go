// Package wikipedia fetches Wikipedia intro extracts (CC BY-SA 4.0) for the
// en, pt and es sitelinks of stored Wikidata records (SDD 06 §7.3).
package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// Langs are the supported Wikipedia editions.
var Langs = []string{"en", "pt", "es"}

// Batch is the TextExtracts cap of exintro pages per request.
const Batch = 20

// Workers caps concurrent batch requests; the per-host limits of httpx still
// bound the rate.
const Workers = 4

var _ ingest.Connector = (*Connector)(nil)

// Connector fetches the extracts of one language edition. A Target carries the
// Wikidata QID in ID and the page title in Type.
type Connector struct {
	lang string
	// APIURL overrides the edition endpoint; tests only.
	APIURL string
}

// New returns the connector of lang, one of Langs.
func New(lang string) (*Connector, error) {
	if !slices.Contains(Langs, lang) {
		return nil, fmt.Errorf("wikipedia: unsupported language %q", lang)
	}

	return &Connector{lang: lang}, nil
}

// Source names the connector in source_records.
func (w *Connector) Source() string { return "wikipedia:" + w.lang }

// Payload is the stored record body.
type Payload struct {
	Qid     string `json:"qid"`
	Lang    string `json:"lang"`
	Title   string `json:"title"`
	Extract string `json:"extract"`
	URL     string `json:"url"`
}

type page struct {
	Title     string            `json:"title"`
	Extract   string            `json:"extract"`
	FullURL   string            `json:"fullurl"`
	PageProps map[string]string `json:"pageprops"`
}

// ParseExtracts keeps the pages whose wikibase_item is one of qids. A redirect
// can land on another entity, a disambiguation page has no usable synopsis,
// and a missing page has neither props nor extract.
func ParseExtracts(raw []byte, qids map[string]bool, lang string) ([]Payload, error) {
	var doc struct {
		Query struct {
			Pages []page `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("decode wikipedia extracts: %w", err)
	}

	var out []Payload

	for _, p := range doc.Query.Pages {
		qid := p.PageProps["wikibase_item"]
		_, disambiguation := p.PageProps["disambiguation"]
		extract := strings.TrimSpace(p.Extract)

		if !qids[qid] || disambiguation || extract == "" {
			continue
		}

		out = append(out, Payload{Qid: qid, Lang: lang, Title: p.Title, Extract: extract, URL: p.FullURL})
	}

	return out, nil
}

// targetsSQL: $1 sitelink key, $2 after, $3 since (null = any), $4 only missing
// (skip QIDs already stored for this edition), $5 source, $6 limit (0 = none).
const targetsSQL = `
SELECT s.external_id, s.payload->'sitelinks'->>$1
FROM source_records s
WHERE s.source = 'wikidata'
  AND s.payload->'sitelinks'->>$1 <> ''
  AND s.external_id > $2
  AND ($3::timestamptz IS NULL OR s.fetched_at >= $3)
  AND (NOT $4::boolean OR NOT EXISTS (
    SELECT 1 FROM source_records w WHERE w.source = $5 AND w.external_id = s.external_id))
ORDER BY s.external_id
LIMIT CASE WHEN $6::int > 0 THEN $6::int END`

// Targets yields the QIDs whose Wikidata record has a sitelink in this
// edition, after the cursor. Bootstrap takes only QIDs without an extract,
// incremental those fetched since p.Since, full all of them.
func (w *Connector) Targets(
	ctx context.Context,
	db database.Executor,
	p ingest.RunParams,
	after string,
) iter.Seq2[ingest.Target, error] {
	return func(yield func(ingest.Target, error) bool) {
		var since any
		if p.Mode == run.ModeIncremental && !p.Since.IsZero() {
			since = p.Since
		}

		rows, err := db.QueryContext(ctx, targetsSQL,
			w.lang+"wiki", after, since, p.Mode == run.ModeBootstrap, w.Source(), p.Limit)
		if err != nil {
			yield(ingest.Target{}, fmt.Errorf("wikipedia targets: %w", err))

			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var qid, title string
			if err := rows.Scan(&qid, &title); err != nil {
				yield(ingest.Target{}, fmt.Errorf("scan wikipedia target: %w", err))

				return
			}

			if !yield(ingest.Target{ID: qid, Type: title}, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(ingest.Target{}, fmt.Errorf("wikipedia targets: %w", err))
		}
	}
}

func (w *Connector) endpoint() string {
	if w.APIURL != "" {
		return w.APIURL
	}

	return "https://" + w.lang + ".wikipedia.org/w/api.php"
}

func (w *Connector) extracts(ctx context.Context, c *ingest.Client, chunk []ingest.Target) ([]Payload, error) {
	titles := make([]string, len(chunk))
	qids := make(map[string]bool, len(chunk))

	for i, t := range chunk {
		titles[i] = t.Type
		qids[t.ID] = true
	}

	q := url.Values{
		"action":        {"query"},
		"format":        {"json"},
		"formatversion": {"2"},
		"redirects":     {"1"},
		"prop":          {"extracts|pageprops|info"},
		"exintro":       {"1"},
		"explaintext":   {"1"},
		"exlimit":       {strconv.Itoa(Batch)},
		"ppprop":        {"wikibase_item|disambiguation"},
		"inprop":        {"url"},
		"titles":        {strings.Join(titles, "|")},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.endpoint()+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build extracts request: %w", err)
	}

	resp, err := c.JSON.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wikipedia %s: %w", w.lang, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikipedia %s: status %d", w.lang, resp.StatusCode)
	}

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode wikipedia %s: %w", w.lang, err)
	}

	return ParseExtracts(raw, qids, w.lang)
}

// Fetch resolves the batch in chunks of Batch titles, Workers requests at a
// time. Records keep the order of the targets; a failed chunk becomes one
// TargetError per target.
func (w *Connector) Fetch(
	ctx context.Context,
	c *ingest.Client,
	batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	chunks := slices.Collect(slices.Chunk(batch, Batch))
	results := make([][]Payload, len(chunks))
	failures := make([]error, len(chunks))

	var wg sync.WaitGroup

	sem := make(chan struct{}, Workers)

	for i, chunk := range chunks {
		sem <- struct{}{}

		wg.Go(func() {
			defer func() { <-sem }()

			results[i], failures[i] = w.extracts(ctx, c, chunk)
		})
	}

	wg.Wait()

	var (
		records []ingest.Record
		errs    []ingest.TargetError
	)

	for i, chunk := range chunks {
		if failures[i] != nil {
			for _, t := range chunk {
				errs = append(errs, ingest.TargetError{Target: t, Err: failures[i]})
			}

			continue
		}

		for _, p := range results[i] {
			body, err := json.Marshal(p)
			if err != nil {
				errs = append(errs, ingest.TargetError{Target: ingest.Target{ID: p.Qid}, Err: err})

				continue
			}

			records = append(records, ingest.Record{Source: w.Source(), ExternalID: p.Qid, Payload: body})
		}
	}

	return records, errs
}
