//go:build integration

package legacyimport_test

import (
	"io"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

func TestEntitiesStage(t *testing.T) {
	t.Parallel()

	_, db := fixtureDB(t)
	ctx := t.Context()

	// Books without usable edition data (01-1): an invalid ISBN, none, and one already held by i3.
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO legacy_import.items (id, type, slug, wikidata_qid, external_ids) VALUES
  ('01900000-0000-7000-8000-000000000006', 'book', 'bad-isbn', 'Q6', '{"isbn13": "123"}'),
  ('01900000-0000-7000-8000-000000000007', 'book', 'no-isbn', 'Q7', '{}'),
  ('01900000-0000-7000-8000-000000000008', 'book', 'same-isbn', 'Q8', '{"isbn13": "9780000000002"}')`,
	); err != nil {
		t.Fatal(err)
	}

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"catalog", "entities"})
	if err != nil {
		t.Fatal(err)
	}

	r := &legacyimport.Runner{DB: db, Batch: 2, Out: io.Discard, Now: time.Now}
	if err := r.Run(ctx, stages); err != nil {
		t.Fatalf("run: %v", err)
	}

	links := `SELECT string_agg(concat_ws(':', ie.role, e.kind, e.default_name, ie.source, e.wikidata_qid), ','
  ORDER BY ie.role, e.default_name)
FROM item_entities ie JOIN entities e ON e.id = ie.entity_id WHERE ie.item_id = $1`

	checks := []struct {
		name, query, arg, want string
	}{
		{
			"strings resolved through QIDs, the rest native", links, "01900000-0000-7000-8000-000000000001",
			"country:country:United States:legacy,developer:organization:Valve:wikidata:Q193559," +
				"genre:genre:shooter:wikidata:Q185029,platform:platform:Windows:legacy",
		},
		{
			"roleless book creator is an author", links, "01900000-0000-7000-8000-000000000003",
			"author:person:Ana Writer:legacy,genre:genre:shooter:legacy",
		},
		{
			"native genre apart from the Wikidata one",
			`SELECT string_agg(slug || ':' || coalesce(wikidata_qid, '-'), ','
  ORDER BY slug) FROM entities WHERE kind = $1`,
			"genre",
			"shooter:Q185029,shooter-2:-",
		},
		// DQ-09 lists these for review.
		{"native entities", `SELECT count(*)::text FROM entities WHERE wikidata_qid IS NULL AND $1 <> ''`, "x", "5"},
		{
			"non-Latin name gets a hashed slug", `SELECT e.kind || ':' || (e.slug LIKE 'person-%')::text FROM entities e
JOIN item_entities ie ON ie.entity_id = e.id WHERE ie.item_id = $1 AND ie.role = 'creator'`,
			"01900000-0000-7000-8000-000000000004", "person:true",
		},
		// DQ-10 lists editions with the defaulted language.
		{
			"one edition per isbn13",
			`SELECT string_agg(concat_ws(':', w.slug, e.isbn13, e.title, e.language, e.source), ',')
FROM editions e JOIN items w ON w.id = e.work_id WHERE $1 <> ''`,
			"x",
			"some-book:9780000000002:Un libro:und:legacy",
		},
		{"checkpoint", `SELECT status || ' ' || rows FROM legacy_import_runs WHERE stage = $1`, "entities", "done 8"},
	}

	for _, c := range checks {
		var got string
		if err := db.QueryRowContext(ctx, c.query, c.arg).Scan(&got); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}

		if got != c.want {
			t.Errorf("%s = %q, want %q", c.name, got, c.want)
		}
	}
}
