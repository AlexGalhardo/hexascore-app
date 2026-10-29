//go:build integration

package catalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestDQ(t *testing.T) {
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	clean, err := catalog.RunDQ(ctx, db)
	if err != nil || clean.Errors != 0 || clean.Warnings != 0 || len(clean.Checks) != 14 {
		t.Fatalf("empty database: %+v err %v", clean, err)
	}

	// The schema already blocks several violations with CHECK/UNIQUE constraints;
	// the asserts are the second line of defence for data loaded around them.
	exec(`DO $$ DECLARE c record; BEGIN
FOR c IN SELECT conrelid::regclass AS t, conname FROM pg_constraint
 WHERE contype IN ('c','u') AND conrelid::regclass::text IN
 ('items','entities','item_translations','entity_translations','editions')
LOOP EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I CASCADE', c.t, c.conname); END LOOP; END $$`)

	id := func(n int) string { return fmt.Sprintf("00000000-0000-7000-8000-%012d", n) }
	exec(`INSERT INTO data_sources (code, name, licence, attribution, url) VALUES
('wikidata','Wikidata','CC0-1.0',NULL,'https://www.wikidata.org'),
('wikipedia','Wikipedia','CC-BY-SA-4.0',NULL,'https://wikipedia.org'),
('tmdb','TMDB','proprietary-tmdb','Powered by TMDB','https://www.themoviedb.org')`)

	item := func(n int, typ, slug, extra string) {
		exec(
			`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1,$2,$3,$4)`,
			id(n),
			typ,
			slug,
			fmt.Sprintf("Q%d", n),
		)
		if extra != "" {
			exec(`UPDATE items SET `+extra+` WHERE id = $1`, id(n))
		}
		if n != 1 { // item 1 has no en translation (DQ-01)
			exec(
				`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1,'en','T','wikidata')`,
				id(n),
			)
		}
	}
	item(
		1,
		"game",
		"bad Slug",
		"",
	) // DQ-01, DQ-04
	item(
		2,
		"book",
		"book-x",
		"",
	) // DQ-10 (no editions)
	item(
		3,
		"game",
		"g3",
		`cover_key='k', cover_source='tmdb', provenance='{"title":"nope"}', release_precision='day'`,
	) // DQ-07, 11, 13
	item(4, "game", "g4", "")
	item(5, "game", "g5", "")
	item(6, "game", "old-slug", "") // DQ-04 (redirect old_slug)
	item(7, "game", "g7", "")
	exec(
		`UPDATE item_translations SET synopsis='x' WHERE item_id = $1`,
		id(7),
	) // DQ-06
	exec(
		`INSERT INTO item_links (from_item_id, to_item_id, relation, source) VALUES ($1,$2,'based_on','wikidata'),($2,$1,'based_on','wikidata')`,
		id(4),
		id(5),
	) // DQ-14

	exec(
		`INSERT INTO editions (id, work_id, isbn13, title, language, source) VALUES ($1,$2,'9780306406158','E','und','wikidata')`,
		id(50),
		id(2),
	) // DQ-02, DQ-10
	exec(
		`INSERT INTO entities (id, kind, slug, default_name, created_at) VALUES ($1,'person','p1','P1', now() - interval '10 days')`,
		id(101),
	) // DQ-03, DQ-09
	exec(
		`INSERT INTO entities (id, kind, slug, default_name, wikidata_qid) VALUES ($1,'person','p2','P2','Q999')`,
		id(102),
	)
	exec(
		`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1,$2,'director','wikidata')`,
		id(7),
		id(102),
	) // DQ-08 (director on a game)
	exec(
		`INSERT INTO entity_stats (entity_id, type, item_count, scored_count, refreshed_at) VALUES ($1,'game',1,0, now() - interval '3 hours')`,
		id(102),
	) // DQ-12
	exec(`INSERT INTO catalog_redirects (scope, namespace, old_slug, target_id, reason) VALUES
('item','game','old-slug',$1,'rename'), ('item','game','gone',$2,'rename')`, id(7), id(900)) // DQ-04, DQ-05

	report, err := catalog.RunDQ(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range report.Checks {
		if c.Count == 0 {
			t.Errorf("%s did not fire on its bad fixture", c.ID)
		}
	}
	if report.Errors == 0 || report.Warnings == 0 {
		t.Errorf("report: errors %d warnings %d", report.Errors, report.Warnings)
	}
}
