package discover_test

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/discover"
)

func TestQueryContinuesBelowThePositionInAStableOrder(t *testing.T) {
	q := discover.Query("movie", 40, "http://www.wikidata.org/entity/Q10", 3, 500)

	for _, want := range []string{
		"wd:Q11424", // film class
		"?links >= 3",
		"?links < 40 || (?links = 40 && STR(?item) > \"http://www.wikidata.org/entity/Q10\")",
		"ORDER BY DESC(?links) ?item LIMIT 500",
	} {
		if !strings.Contains(q, want) {
			t.Errorf("query lacks %q:\n%s", want, q)
		}
	}
}

func TestParseRowsReadsQidsAndSitelinks(t *testing.T) {
	rows, err := discover.ParseRows([]byte(`{"results":{"bindings":[
{"item":{"value":"http://www.wikidata.org/entity/Q83495"},"links":{"value":"61"}},
{"item":{"value":"http://www.wikidata.org/entity/Q4"},"links":{"value":"58"}}]}}`))
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 2 || rows[0].QID != "Q83495" || rows[0].Links != 61 ||
		rows[1].IRI != "http://www.wikidata.org/entity/Q4" {
		t.Errorf("rows = %+v", rows)
	}

	if _, err := discover.ParseRows(
		[]byte(`{"results":{"bindings":[{"item":{"value":"x/Q1"},"links":{"value":"many"}}]}}`),
	); err == nil {
		t.Error("a non-numeric sitelink count must be an error")
	}
}
