// Package openlibrary is the Open Library connector (SDD 06 §7.3): page count, cover ids, the description and the
// edition's ISBN, language and year for books with an ISBN-13 or an Open Library id. Port of legacy connectors/openlibrary.py.
//
// Terms: the Internet Archive asserts no new copyright or other rights over the Open Library database
// (https://openlibrary.org/developers/licensing, read 2026-11-01; owner cleared A-05 in Q-A16). The description
// is only the input of the synopsis step and never reaches item_translations (FR-06-3).
package openlibrary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// BaseURL is the Open Library site; its JSON documents live under the same paths as the pages.
const BaseURL = "https://openlibrary.org"

// MaxCovers caps the cover ids kept per record.
const MaxCovers = 3

var (
	_ ingest.Connector = (*Connector)(nil)

	editionID = regexp.MustCompile(`^OL\d+M$`)
	workID    = regexp.MustCompile(`^OL\d+W$`)
	workKey   = regexp.MustCompile(`^/works/OL\d+W$`)
	notISBN   = regexp.MustCompile(`[^0-9X]`)
	yearRe    = regexp.MustCompile(`\b(1[4-9]|20)\d{2}\b`)
)

// Payload is the stored record body; nil Pages and Overview are stored as null like the legacy records. ISBN (raw,
// ISBN-13 first), Language (MARC code) and Year describe the edition and stay empty for a work document.
type Payload struct {
	Qid      string  `json:"qid"`
	Key      string  `json:"key"`
	Pages    *int    `json:"pages"`
	Covers   []int   `json:"covers"`
	Overview *string `json:"overview"`
	ISBN     string  `json:"isbn,omitempty"`
	Language string  `json:"language,omitempty"`
	Year     int     `json:"year,omitempty"`
}

type doc struct {
	Key           string                 `json:"key"`
	NumberOfPages int                    `json:"number_of_pages"`
	Covers        []int                  `json:"covers"`
	Works         []struct{ Key string } `json:"works"`
	Description   json.RawMessage        `json:"description"`
	ISBN13        []string               `json:"isbn_13"`
	ISBN10        []string               `json:"isbn_10"`
	Languages     []struct{ Key string } `json:"languages"`
	PublishDate   string                 `json:"publish_date"`
}

func decode(raw []byte) (doc, error) {
	var d doc
	if err := json.Unmarshal(raw, &d); err != nil {
		return doc{}, fmt.Errorf("decode open library document: %w", err)
	}

	return d, nil
}

// coverIDs drops -1, Open Library's "no cover" placeholder.
func (d doc) coverIDs() []int {
	var out []int

	for _, c := range d.Covers {
		if c > 0 {
			out = append(out, c)
		}
	}

	return out
}

// overview reads a description given as a plain string or as a typed {"type": "/type/text", "value": ...}.
func (d doc) overview() *string {
	var text string
	if json.Unmarshal(d.Description, &text) != nil {
		var typed struct {
			Value string `json:"value"`
		}

		_ = json.Unmarshal(d.Description, &typed)
		text = typed.Value
	}

	if text = strings.TrimSpace(text); text == "" {
		return nil
	}

	return &text
}

// DocumentPath picks the document to fetch: edition data (ISBN or OL…M) carries the page count, so it wins over a
// work (OL…W). It returns "" when no usable id exists.
func DocumentPath(externalIDs map[string]string) string {
	isbn := notISBN.ReplaceAllString(strings.ToUpper(externalIDs["isbn13"]), "")
	olid := externalIDs["openlibrary"]

	switch {
	case len(isbn) == 13:
		return "/isbn/" + isbn + ".json"
	case editionID.MatchString(olid):
		return "/books/" + olid + ".json"
	case workID.MatchString(olid):
		return "/works/" + olid + ".json"
	default:
		return ""
	}
}

// WorkPath is the path of the work an edition belongs to; "" for a work document or an unexpected key.
func WorkPath(raw []byte) string {
	d, err := decode(raw)
	if err != nil || len(d.Works) == 0 || !workKey.MatchString(d.Works[0].Key) {
		return ""
	}

	return d.Works[0].Key + ".json"
}

// ParseDoc builds the payload of one document; work (the edition's work document, or nil) fills the cover and the
// description the edition lacks.
func ParseDoc(raw []byte, qid string, work []byte) (Payload, error) {
	d, err := decode(raw)
	if err != nil {
		return Payload{}, err
	}

	p := Payload{Qid: qid, Key: d.Key, Covers: d.coverIDs(), Overview: d.overview()}
	if d.NumberOfPages > 0 {
		p.Pages = &d.NumberOfPages
	}

	if strings.HasPrefix(d.Key, "/books/") {
		p.ISBN = firstOf(d.ISBN13, d.ISBN10)
		if len(d.Languages) > 0 {
			p.Language = strings.TrimPrefix(d.Languages[0].Key, "/languages/")
		}

		// publish_date is free text ("1984", "October 1, 1988"); the first plausible year is kept.
		if y := yearRe.FindString(d.PublishDate); y != "" {
			p.Year, _ = strconv.Atoi(y)
		}
	}

	if work != nil {
		parent, err := decode(work)
		if err != nil {
			return Payload{}, err
		}

		if len(p.Covers) == 0 {
			p.Covers = parent.coverIDs()
		}

		if p.Overview == nil {
			p.Overview = parent.overview()
		}
	}

	if p.Covers == nil {
		p.Covers = []int{}
	}

	p.Covers = p.Covers[:min(MaxCovers, len(p.Covers))]

	return p, nil
}

func firstOf(lists ...[]string) string {
	for _, l := range lists {
		if len(l) > 0 {
			return l[0]
		}
	}

	return ""
}

// Connector fetches Open Library documents for stored books. A Target carries the QID in ID and Ref(...) in Type.
type Connector struct {
	// BaseURL overrides BaseURL; tests only.
	BaseURL string
}

// Source names the connector in source_records.
func (*Connector) Source() string { return "openlibrary" }

// Ref packs the document path and whether the item already has an English Wikipedia synopsis.
func Ref(path string, hasSynopsis bool) string {
	if hasSynopsis {
		return path + "|1"
	}

	return path + "|0"
}

// targetsSQL: $1 after, $2 since (null = any), $3 only missing, $4 limit (0 = none), $5 QIDs (empty = all).
// Only missing skips records in the current shape (an overview key) and older ones with nothing left to fill (legacy complete_ids).
const targetsSQL = `
SELECT i.wikidata_qid, i.external_ids, EXISTS (SELECT 1 FROM source_records w
  WHERE w.source = 'wikipedia:en' AND w.external_id = i.wikidata_qid)
FROM items i
WHERE i.type = 'book' AND (i.external_ids ? 'isbn13' OR i.external_ids ? 'openlibrary')
  AND i.wikidata_qid > $1
  AND (coalesce(cardinality($5::text[]), 0) = 0 OR i.wikidata_qid = ANY($5::text[]))
  AND ($2::timestamptz IS NULL OR i.updated_at >= $2)
  AND (NOT $3::boolean OR NOT EXISTS (SELECT 1 FROM source_records o
    WHERE o.source = 'openlibrary' AND o.external_id = i.wikidata_qid AND (o.payload ? 'overview'
      OR (o.payload #>> '{covers,0}' IS NOT NULL AND EXISTS (SELECT 1 FROM source_records w
        WHERE w.source = 'wikipedia:en' AND w.external_id = o.external_id)))))
ORDER BY i.wikidata_qid
LIMIT CASE WHEN $4::int > 0 THEN $4::int END`

// Targets yields the books with a usable ISBN-13 or Open Library id, after the cursor.
func (*Connector) Targets(
	ctx context.Context, db database.Executor, p ingest.RunParams, after string,
) iter.Seq2[ingest.Target, error] {
	return func(yield func(ingest.Target, error) bool) {
		var since any
		if p.Mode == run.ModeIncremental && !p.Since.IsZero() {
			since = p.Since
		}

		rows, err := db.QueryContext(ctx, targetsSQL, after, since, p.Mode == run.ModeBootstrap, p.Limit, p.QIDs)
		if err != nil {
			yield(ingest.Target{}, fmt.Errorf("openlibrary targets: %w", err))

			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var (
				qid         string
				raw         []byte
				hasSynopsis bool
			)
			if err := rows.Scan(&qid, &raw, &hasSynopsis); err != nil {
				yield(ingest.Target{}, fmt.Errorf("scan openlibrary target: %w", err))

				return
			}

			var ids map[string]string
			if json.Unmarshal(raw, &ids) != nil {
				continue
			}

			path := DocumentPath(ids)
			if path == "" {
				continue
			}

			if !yield(ingest.Target{ID: qid, Type: Ref(path, hasSynopsis)}, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(ingest.Target{}, fmt.Errorf("openlibrary targets: %w", err))
		}
	}
}

// get returns the body of a document; nil without error for a 404.
func (c *Connector) get(ctx context.Context, client *ingest.Client, path string) ([]byte, error) {
	base := c.BaseURL
	if base == "" {
		base = BaseURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build open library request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.JSON.Do(req)
	if err != nil {
		return nil, fmt.Errorf("open library %s: %w", path, err)
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, fmt.Errorf("open library %s: status %d", path, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read open library %s: %w", path, err)
	}

	return body, nil
}

func (c *Connector) one(ctx context.Context, client *ingest.Client, t ingest.Target) (*ingest.Record, error) {
	path, flag, _ := strings.Cut(t.Type, "|")

	raw, err := c.get(ctx, client, path)
	if err != nil || raw == nil {
		return nil, err
	}

	d, err := decode(raw)
	if err != nil {
		return nil, err
	}

	var work []byte

	// The second request is only spent when the edition leaves something to fill.
	if wp := WorkPath(raw); wp != "" && (len(d.coverIDs()) == 0 || (flag != "1" && d.overview() == nil)) {
		if work, err = c.get(ctx, client, wp); err != nil {
			return nil, err
		}
	}

	p, err := ParseDoc(raw, t.ID, work)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("encode open library %s: %w", t.ID, err)
	}

	return &ingest.Record{Source: c.Source(), ExternalID: t.ID, Payload: body}, nil
}

// Fetch resolves the batch one target at a time (the openlibrary.org budget is 2.5/s, SDD 06 §7.2). A document
// Open Library does not know is skipped; any other failure is kept for the dead-letter table.
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
