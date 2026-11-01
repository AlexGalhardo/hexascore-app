package normalize

import (
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/editions"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// ExtractEditions lists the editions of a book from its Wikidata record: ISBN-13 (P212) and ISBN-10 (P957) on the
// work, and the edition items it links (P747, by QID only: their language, publisher and year need the edition
// entities, which are not fetched). It returns the selected candidates and how many invalid ISBNs were dropped.
func ExtractEditions(qid string, recs []store.Record) ([]editions.Candidate, int, error) {
	wd, _, err := decode(qid, recs)
	if err != nil {
		return nil, 0, err
	}

	if wd.Type != "book" {
		return nil, 0, nil
	}

	var in []editions.Candidate

	for _, prop := range []string{"P212", "P957"} {
		for _, isbn := range values(wd, prop) {
			in = append(
				in,
				editions.Candidate{ISBN: isbn, Title: wd.Labels["en"], Language: "und", Source: wikidataSource},
			)
		}
	}

	for _, ed := range values(wd, "P747") {
		in = append(in, editions.Candidate{QID: ed, Title: wd.Labels["en"], Language: "und", Source: wikidataSource})
	}

	out, dropped := editions.Select(in, editions.MaxPerWork)

	return out, dropped, nil
}
