package normalize

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/editions"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
)

// editionLanguages maps the P407 language items to editions.language (DQ-10); anything else is "und".
// Brazilian Portuguese counts as "pt", like MARC "por".
var editionLanguages = map[string]string{"Q5146": "pt", "Q750553": "pt", "Q1860": "en", "Q1321": "es"}

// ExtractEditions lists the editions of a book from its Wikidata record: ISBN-13 (P212) and ISBN-10 (P957) on the
// work, and the edition items it links (P747). An edition whose entity record (source wikidata:entity) is among recs
// takes its language (P407), year (P577), ISBN and label from it; without one it is kept by QID only. It returns the
// selected candidates and how many invalid ISBNs were dropped.
func ExtractEditions(qid string, recs []store.Record) ([]editions.Candidate, int, error) {
	wd, _, err := decode(qid, recs)
	if err != nil {
		return nil, 0, err
	}

	if wd.Type != "book" {
		return nil, 0, nil
	}

	entities := map[string]wikidata.Payload{}

	for _, r := range recs {
		if r.Source != wikidata.EntitySource {
			continue
		}

		var e wikidata.EntityPayload
		if err := json.Unmarshal(r.Payload, &e); err != nil {
			return nil, 0, fmt.Errorf("edition %s: %w", r.ExternalID, err)
		}

		entities[r.ExternalID] = wikidata.Payload{Labels: e.Labels, Claims: e.Claims}
	}

	var in []editions.Candidate

	// Editions go first, so an ISBN they share with the work keeps the edition's QID and details.
	for _, ed := range values(wd, "P747") {
		c := editions.Candidate{QID: ed, Title: wd.Labels["en"], Language: "und", Source: wikidataSource}

		if e, ok := entities[ed]; ok {
			c.Title = cmp.Or(e.Labels["en"], c.Title)
			if langs := values(e, "P407"); len(langs) > 0 {
				c.Language = cmp.Or(editionLanguages[langs[0]], "und")
			}

			if d, ok := releaseDate(e); ok {
				c.Year = d.Time.Year()
			}

			c.ISBN = cmp.Or(firstString(e, "P212"), firstString(e, "P957"))
		}

		in = append(in, c)
	}

	for _, prop := range []string{"P212", "P957"} {
		for _, isbn := range values(wd, prop) {
			in = append(
				in,
				editions.Candidate{ISBN: isbn, Title: wd.Labels["en"], Language: "und", Source: wikidataSource},
			)
		}
	}

	out, dropped := editions.Select(in, editions.MaxPerWork)

	return out, dropped, nil
}

// rowQuerier is the part of *database.Tx WithEditionRecords needs.
type rowQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// WithEditionRecords appends to recs the stored wikidata:entity records of the book's P747 editions, so
// ExtractEditions can read their details. Editions without a stored record are simply absent.
func WithEditionRecords(ctx context.Context, q rowQuerier, qid string, recs []store.Record) ([]store.Record, error) {
	wd, _, err := decode(qid, recs)
	if err != nil {
		return nil, err
	}

	eds := values(wd, "P747")
	if wd.Type != "book" || len(eds) == 0 {
		return recs, nil
	}

	// QIDs come from values(), which keeps only entity ids, so the array literal needs no escaping.
	rows, err := q.QueryContext(ctx, `SELECT external_id, payload FROM source_records
WHERE source = $1 AND external_id = ANY($2::text[])`, wikidata.EntitySource, "{"+strings.Join(eds, ",")+"}")
	if err != nil {
		return nil, fmt.Errorf("edition records of %s: %w", qid, err)
	}
	defer func() { _ = rows.Close() }()

	out := slices.Clip(recs)

	for rows.Next() {
		r := store.Record{Source: wikidata.EntitySource}
		if err := rows.Scan(&r.ExternalID, &r.Payload); err != nil {
			return nil, fmt.Errorf("edition records of %s: %w", qid, err)
		}

		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("edition records of %s: %w", qid, err)
	}

	return out, nil
}
