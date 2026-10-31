package wikidata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// APIURL is the Wikidata action API.
const APIURL = "https://www.wikidata.org/w/api.php"

// Batch is the wbgetentities page size.
const Batch = 50

var _ ingest.Connector = (*Connector)(nil)

// refs remembers the referenced entities already seen in this process, so
// a person credited in many items is fetched and stored once.
type refs struct {
	mu    sync.Mutex
	label map[string]string
}

func (r *refs) get(q string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	l, ok := r.label[q]

	return l, ok
}

func (r *refs) put(q, label string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.label[q] = label
}

// getEntities fetches one page of entities. The result keeps one raw entity
// per requested id that the API returned.
func (w *Connector) getEntities(
	ctx context.Context,
	c *ingest.Client,
	ids []string,
) (map[string]json.RawMessage, error) {
	endpoint := w.APIURL
	if endpoint == "" {
		endpoint = APIURL
	}

	q := url.Values{
		"action":    {"wbgetentities"},
		"ids":       {strings.Join(ids, "|")},
		"props":     {"labels|aliases|claims|sitelinks"},
		"languages": {strings.Join(LabelLangs, "|")},
		"format":    {"json"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build wbgetentities request: %w", err)
	}

	resp, err := c.JSON.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wbgetentities: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wbgetentities: status %d", resp.StatusCode)
	}

	var doc struct {
		Entities map[string]json.RawMessage `json:"entities"`
		Error    *struct {
			Info string `json:"info"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode wbgetentities: %w", err)
	}

	if doc.Error != nil {
		return nil, fmt.Errorf("wbgetentities: %s", doc.Error.Info)
	}

	return doc.Entities, nil
}

// Fetch resolves a batch of targets into "wikidata" records plus one
// "wikidata:entity" record per newly referenced QID. Missing entities are
// skipped; an undecodable one or a failed page becomes a TargetError.
func (w *Connector) Fetch(
	ctx context.Context,
	c *ingest.Client,
	batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	var (
		records []ingest.Record
		errs    []ingest.TargetError
	)

	for chunk := range slices.Chunk(batch, Batch) {
		ids := make([]string, len(chunk))
		for i, t := range chunk {
			ids[i] = t.ID
		}

		raws, err := w.getEntities(ctx, c, ids)
		if err != nil {
			for _, t := range chunk {
				errs = append(errs, ingest.TargetError{Target: t, Err: err})
			}

			continue
		}

		for _, t := range chunk {
			raw, ok := raws[t.ID]
			if !ok {
				continue
			}

			e, err := ParseEntity(raw)
			if err != nil {
				errs = append(errs, ingest.TargetError{Target: t, Err: err})

				continue
			}

			if e.Missing != nil {
				continue
			}

			p := Trim(e, t.Type)

			entityRecs, err := w.resolveRefs(ctx, c, &p)
			if err != nil {
				errs = append(errs, ingest.TargetError{Target: t, Err: err})

				continue
			}

			body, err := json.Marshal(p)
			if err != nil {
				errs = append(errs, ingest.TargetError{Target: t, Err: fmt.Errorf("encode %s: %w", t.ID, err)})

				continue
			}

			records = append(records, ingest.Record{Source: w.Source(), ExternalID: e.ID, Payload: body})
			records = append(records, entityRecs...)
		}
	}

	return records, errs
}

// resolveRefs fills p.Refs with English (else "mul") labels and returns
// records for the entities not seen before.
func (w *Connector) resolveRefs(ctx context.Context, c *ingest.Client, p *Payload) ([]ingest.Record, error) {
	qids := ReferencedQids(*p)

	var need []string

	for _, q := range qids {
		if _, seen := w.refs.get(q); !seen {
			need = append(need, q)
		}
	}

	var out []ingest.Record

	for chunk := range slices.Chunk(need, Batch) {
		raws, err := w.getEntities(ctx, c, chunk)
		if err != nil {
			return nil, err
		}

		for _, q := range chunk {
			label := ""

			if raw, ok := raws[q]; ok {
				e, err := ParseEntity(raw)
				if err != nil {
					return nil, err
				}

				if e.Missing == nil {
					label = refLabel(e)

					body, err := json.Marshal(TrimEntity(e))
					if err != nil {
						return nil, fmt.Errorf("encode entity %s: %w", q, err)
					}

					out = append(out, ingest.Record{Source: "wikidata:entity", ExternalID: q, Payload: body})
				}
			}

			w.refs.put(q, label)
		}
	}

	for _, q := range qids {
		if l, _ := w.refs.get(q); l != "" {
			p.Refs[q] = l
		}
	}

	return out, nil
}
