//go:build integration

package legacyimport_test

import (
	"io"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

func TestCatalogStage(t *testing.T) {
	t.Parallel()

	_, db := fixtureDB(t)
	ctx := t.Context()

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"catalog"})
	if err != nil {
		t.Fatal(err)
	}

	// Batches of two items: the five fixture items take three batches.
	r := &legacyimport.Runner{DB: db, Batch: 2, Out: io.Discard, Now: time.Now}
	if err := r.Run(ctx, stages); err != nil {
		t.Fatalf("run: %v", err)
	}

	checks := []struct{ name, query, want string }{
		{"ids preserved", `SELECT count(*)::text FROM items i JOIN legacy_import.items l USING (id)
WHERE i.type = l.type AND i.slug = l.slug AND i.wikidata_qid = l.wikidata_qid AND i.external_ids = l.external_ids
  AND i.created_at = l.created_at AND i.updated_at = l.updated_at`, "5"},
		{
			"release and visibility",
			`SELECT string_agg(slug || ':' || release_precision || '/' || release_status || '/' || visibility, ',' ORDER BY id)
FROM items`,
			"half-life:day/released/public,future-film:day/announced/public,some-book:unknown/announced/public," +
				"some-series:day/released/public,bare-game:day/released/public",
		},
		{"popularity prior", `SELECT popularity_prior::text FROM items WHERE slug = 'half-life'`, "12.5"},
		{
			"scores left to the recompute",
			`SELECT count(*)::text FROM items WHERE score IS NOT NULL OR rating_count <> 0`,
			"0",
		},
		{
			"typed columns",
			`SELECT string_agg(concat_ws('/', slug, runtime_minutes, page_count, season_count, episode_count,
  jsonb_array_length(trailers)), ',' ORDER BY id) FROM items`,
			"half-life/1,future-film/136/0,some-book/300/0,some-series/2/20/0,bare-game/0",
		},
		{
			"cover keys",
			`SELECT concat_ws('|', cover_key, cover_thumb_key, cover_attribution->>'source') FROM items WHERE slug = 'half-life'`,
			"covers/01900000-0000-7000-8000-000000000001-0123456789abcdef.webp|" +
				"covers/01900000-0000-7000-8000-000000000001-0123456789abcdef-t.webp|IGDB",
		},
		{
			"non-Blob cover gets no key",
			`SELECT count(*)::text FROM items WHERE slug = 'future-film' AND cover_key IS NULL`,
			"1",
		},
		{
			"translations",
			`SELECT string_agg(concat_ws('/', locale, title_source, synopsis_source, synopsis_licence), ',' ORDER BY locale)
FROM item_translations WHERE item_id = '01900000-0000-7000-8000-000000000001'`,
			"en/wikidata/wikipedia/CC BY-SA 4.0,pt-BR/wikidata/hexascore/proprietary",
		},
		{"translation count", `SELECT count(*)::text FROM item_translations`, "6"},
		{
			"affiliate links",
			`SELECT store || ' ' || url FROM affiliate_links`,
			"steam https://store.steampowered.com/app/70",
		},
		{
			"source records",
			`SELECT string_agg(s.source || ':' || s.external_id || ':' || coalesce(s.item_id::text, '-'), ',' ORDER BY s.source, s.external_id)
FROM source_records s JOIN legacy_import.source_records l USING (source, external_id)
WHERE s.payload = l.payload AND s.hash = l.hash AND s.normalized_at IS NOT DISTINCT FROM l.normalized_at`,
			"steam:Q279744:01900000-0000-7000-8000-000000000001," +
				"wikidata:Q279744:01900000-0000-7000-8000-000000000001,wikidata:Q999:-",
		},
		// D-09: external scores stay in the legacy dump only.
		{"no external_scores", `SELECT coalesce(to_regclass('public.external_scores')::text, 'none')`, "none"},
		{
			"checkpoint",
			`SELECT status || ' ' || last_key || ' ' || rows FROM legacy_import_runs WHERE stage = 'catalog'`,
			"done 01900000-0000-7000-8000-000000000005 5",
		},
	}

	for _, c := range checks {
		var got string
		if err := db.QueryRowContext(ctx, c.query).Scan(&got); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}

		if got != c.want {
			t.Errorf("%s = %q, want %q", c.name, got, c.want)
		}
	}

	// A crash between a batch commit and the next one replays nothing twice.
	if _, err := db.ExecContext(ctx, `UPDATE legacy_import_runs SET status = 'running',
  last_key = '01900000-0000-7000-8000-000000000002' WHERE stage = 'catalog'`); err != nil {
		t.Fatal(err)
	}

	if err := r.Run(ctx, stages); err != nil {
		t.Fatalf("resume: %v", err)
	}

	var items, records int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM items), (SELECT count(*) FROM source_records)`).
		Scan(&items, &records); err != nil {
		t.Fatal(err)
	}

	if items != 5 || records != 3 {
		t.Fatalf("after replay: items=%d records=%d, want 5 3", items, records)
	}
}

// An out-of-range runtime or page count becomes NULL instead of aborting the batch on the items checks.
func TestCatalogStageDropsOutOfRangeNumbers(t *testing.T) {
	t.Parallel()

	_, db := fixtureDB(t)
	ctx := t.Context()

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO legacy_import.items (id, type, slug, wikidata_qid, external_ids, meta) VALUES
  ('01900000-0000-7000-8000-000000000009', 'movie', 'zero-runtime', 'Q9', '{}', '{"runtimeMinutes": 0}'),
  ('01900000-0000-7000-8000-00000000000a', 'book', 'huge-book', 'Q10', '{}', '{"pages": 99999}')`,
	); err != nil {
		t.Fatal(err)
	}

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"catalog"})
	if err != nil {
		t.Fatal(err)
	}

	r := &legacyimport.Runner{DB: db, Batch: 100, Out: io.Discard, Now: time.Now}
	if err := r.Run(ctx, stages); err != nil {
		t.Fatalf("run: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM items
WHERE slug IN ('zero-runtime', 'huge-book') AND runtime_minutes IS NULL AND page_count IS NULL`).Scan(&n); err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Fatalf("imported with NULL numbers = %d, want 2", n)
	}
}
