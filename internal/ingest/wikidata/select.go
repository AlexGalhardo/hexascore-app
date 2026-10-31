// Package wikidata is the Wikidata connector: SPARQL selects QIDs and
// wbgetentities returns the entities (SDD 06 §7.3). Port of legacy
// connectors/wikidata.py.
package wikidata

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// SparqlURL is the Wikidata Query Service endpoint.
const SparqlURL = "https://query.wikidata.org/sparql"

// IncrementalMinSitelinks keeps long-tail items out of incremental runs.
const IncrementalMinSitelinks = 15

// Types is the selection order.
var Types = []string{"movie", "series", "game", "book"}

// TypeClasses maps an item type to its P31 classes.
var TypeClasses = map[string][]string{
	"movie":  {"Q11424", "Q24869", "Q202866"},
	"series": {"Q5398426", "Q581714", "Q1259759", "Q526877"},
	"game":   {"Q7889"},
	"book":   {"Q7725634", "Q571", "Q8261"},
}

// BootstrapTargets is the top-N per type of a bootstrap run.
var BootstrapTargets = map[string]int{"movie": 35_000, "series": 15_000, "game": 25_000, "book": 25_000}

// Connector selects and fetches Wikidata entities.
type Connector struct {
	// The interface gives Targets no client, so the connector holds it.
	client *ingest.Client
	// SparqlURL overrides SparqlURL; tests only.
	SparqlURL string
	// APIURL overrides APIURL; tests only.
	APIURL string

	refs *refs
}

// New returns a Connector for the public endpoint.
func New(c *ingest.Client) *Connector {
	return &Connector{client: c, SparqlURL: SparqlURL, APIURL: APIURL, refs: &refs{label: map[string]string{}}}
}

// Source names the connector in source_records.
func (*Connector) Source() string { return "wikidata" }

func values(itemType string) string {
	classes := TypeClasses[itemType]
	parts := make([]string, len(classes))

	for i, c := range classes {
		parts[i] = "wd:" + c
	}

	return strings.Join(parts, " ")
}

// TopQuery selects the items with the most sitelinks.
func TopQuery(itemType string, limit int) string {
	return fmt.Sprintf(`SELECT DISTINCT ?item ?links WHERE {
  VALUES ?class { %s }
  ?item wdt:P31 ?class ; wikibase:sitelinks ?links .
} ORDER BY DESC(?links) LIMIT %d`, values(itemType), limit)
}

// ModifiedQuery selects items edited since a time with enough sitelinks.
func ModifiedQuery(itemType string, since time.Time, minSitelinks int) string {
	return fmt.Sprintf(`SELECT DISTINCT ?item WHERE {
  VALUES ?class { %s }
  ?item wdt:P31 ?class ; wikibase:sitelinks ?links ; schema:dateModified ?mod .
  FILTER(?links >= %d && ?mod >= "%s"^^xsd:dateTime)
}`, values(itemType), minSitelinks, since.UTC().Format("2006-01-02T15:04:05Z"))
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

// SelectQids runs the bootstrap query (zero since) or the incremental one.
// A zero limit means the bootstrap target of the type.
func (w *Connector) SelectQids(
	ctx context.Context,
	c *ingest.Client,
	itemType string,
	since time.Time,
	limit int,
) ([]string, error) {
	if _, ok := TypeClasses[itemType]; !ok {
		return nil, fmt.Errorf("wikidata: unknown item type %q", itemType)
	}

	var query string

	if since.IsZero() {
		if limit == 0 {
			limit = BootstrapTargets[itemType]
		}

		query = TopQuery(itemType, limit)
	} else {
		query = ModifiedQuery(itemType, since, IncrementalMinSitelinks)
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
		}
	}

	endpoint := w.SparqlURL
	if endpoint == "" {
		endpoint = SparqlURL
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint+"?"+url.Values{"query": {query}, "format": {"json"}}.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build sparql request: %w", err)
	}

	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := c.JSON.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sparql %s: %w", itemType, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sparql %s: status %d", itemType, resp.StatusCode)
	}

	var out sparqlResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode sparql %s: %w", itemType, err)
	}

	qids := make([]string, 0, len(out.Results.Bindings))

	for _, b := range out.Results.Bindings {
		v := b.Item.Value
		qids = append(qids, v[strings.LastIndex(v, "/")+1:])
	}

	return qids, nil
}

// Targets yields the selected QIDs per type. A bootstrap run (or an
// incremental one without a since time) takes the top items; after skips
// everything up to and including that QID so a continuation resumes. The
// selection is re-run on resume, which is cheap (one query per type).
func (w *Connector) Targets(
	ctx context.Context,
	_ database.Executor,
	p ingest.RunParams,
	after string,
) iter.Seq2[ingest.Target, error] {
	types := p.Types
	if len(types) == 0 {
		types = Types
	}

	since := p.Since
	if p.Mode == run.ModeBootstrap {
		since = time.Time{}
	}

	return func(yield func(ingest.Target, error) bool) {
		skipping := after != ""

		for _, itemType := range types {
			qids, err := w.SelectQids(ctx, w.client, itemType, since, p.Limit)
			if err != nil {
				yield(ingest.Target{}, err)

				return
			}

			for _, q := range qids {
				if skipping {
					skipping = q != after

					continue
				}

				if !yield(ingest.Target{ID: q, Type: itemType}, nil) {
					return
				}
			}
		}
	}
}
