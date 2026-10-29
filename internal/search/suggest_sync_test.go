//go:build integration

package search_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/database/migrations"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

func migratedDB(t *testing.T) *database.DB {
	t.Helper()

	ctx := context.Background()

	db, err := database.Open(ctx, pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	m, err := schema.NewMigrator(db, migrations.All())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx); err != nil {
		t.Fatal(err)
	}

	return db
}

func exec(t *testing.T, db *database.DB, q string, args ...any) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), q, args...); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

// seedItems inserts n public games with 3 translations and 1 alternative title each (4 suggest rows per item)
// and 2 persons with 1 translation (2 rows each).
func seedItems(t *testing.T, db *database.DB, n int) {
	t.Helper()

	exec(
		t,
		db,
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x') ON CONFLICT DO NOTHING`,
	)

	for i := range n {
		id := uuid.NewV7().String()
		exec(
			t,
			db,
			`INSERT INTO items (id, type, slug, wikidata_qid, release_date) VALUES ($1, 'game', $2, $3, '2017-03-03')`,
			id,
			fmt.Sprintf("zelda-%d", i),
			fmt.Sprintf("Q%d", 100+i),
		)

		for _, loc := range []string{"en", "pt-BR", "es"} {
			exec(
				t,
				db,
				`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, $2, $3, 'hexascore')`,
				id,
				loc,
				fmt.Sprintf("Zélda %s %d", loc, i),
			)
		}
		exec(t, db, `INSERT INTO item_alt_titles (item_id, title, kind, source) VALUES ($1, $2, 'alias', 'hexascore')`,
			id, fmt.Sprintf("Zelda alias %d", i))
	}

	for i := range 2 {
		id := uuid.NewV7().String()
		exec(t, db, `INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'person', $2, $3)`,
			id, fmt.Sprintf("link-%d", i), fmt.Sprintf("Link %d", i))
		exec(
			t,
			db,
			`INSERT INTO entity_translations (entity_id, locale, name, name_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id,
			fmt.Sprintf("Link %d", i),
		)
	}
}

func count(t *testing.T, db *database.DB, q string) int {
	t.Helper()

	var n int
	if err := db.QueryRowContext(context.Background(), q).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func TestSuggestSyncConcurrentRunners(t *testing.T) {
	db := migratedDB(t)
	const items = 60
	seedItems(t, db, items)

	if got := count(t, db, `SELECT count(*) FROM search_dirty`); got != items+2 {
		t.Fatalf("dirty rows = %d, want %d", got, items+2)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			// A small batch makes the runners interleave on the queue.
			for {
				n, err := search.SyncSuggest(context.Background(), db, 7)
				if err != nil {
					errs <- err
					return
				}
				if n == 0 {
					return
				}
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}

	// Entity rows: 'Link N' from default_name ('*') and the en translation, distinct locales.
	if got, want := count(t, db, `SELECT count(*) FROM search_suggest`), items*4+2*2; got != want {
		t.Fatalf("suggest rows = %d, want %d", got, want)
	}
	if got := count(t, db, `SELECT count(*) FROM search_dirty`); got != 0 {
		t.Fatalf("dirty rows left = %d", got)
	}
	if got := count(
		t,
		db,
		`SELECT count(*) FROM (SELECT ref_id, locale, label_norm FROM search_suggest GROUP BY 1, 2, 3 HAVING count(*) > 1) d`,
	); got != 0 {
		t.Fatalf("duplicate rows: %d", got)
	}
	if got := count(t, db, `SELECT count(DISTINCT ref_id) FROM search_suggest`); got != items+2 {
		t.Fatalf("missing refs: %d, want %d", got, items+2)
	}

	// Changes and removals follow: hide one item, delete another.
	exec(t, db, `UPDATE items SET visibility = 'hidden' WHERE slug = 'zelda-0'`)
	exec(t, db, `DELETE FROM items WHERE slug = 'zelda-1'`)
	if n, err := search.SyncSuggest(context.Background(), db, search.SuggestBatch); err != nil || n == 0 {
		t.Fatalf("resync: %d, %v", n, err)
	}
	if got := count(t, db, `SELECT count(DISTINCT ref_id) FROM search_suggest`); got != items+2-2 {
		t.Fatalf("refs after hide and delete = %d, want %d", got, items)
	}
}

func TestSuggestPrefixUsesSPGiST(t *testing.T) {
	db := migratedDB(t)
	seedItems(t, db, 20)
	if err := search.DrainSuggest(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	exec(t, db, `ANALYZE search_suggest`)

	// Tiny tables prefer a sequential scan; forbid it to prove the index serves the operator.
	tx, err := db.SQL().BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`SET LOCAL enable_seqscan = off`); err != nil {
		t.Fatal(err)
	}
	rows, err := tx.Query(`EXPLAIN SELECT * FROM search_suggest WHERE label_norm ^@ 'zel'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		plan.WriteString(line + "\n")
	}
	if !strings.Contains(plan.String(), "Index Scan using search_suggest_prefix") &&
		!strings.Contains(plan.String(), "Bitmap Index Scan on search_suggest_prefix") {
		t.Fatalf("plan does not use search_suggest_prefix:\n%s", plan.String())
	}
}
