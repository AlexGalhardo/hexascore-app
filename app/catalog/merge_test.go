//go:build integration

package catalog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func mergeEnv(t *testing.T) (*database.DB, *queue.Queue) {
	t.Helper()
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	q, err := app.Queue()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `INSERT INTO data_sources (code, name, licence, url)
VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`); err != nil {
		t.Fatal(err)
	}
	return db, q
}

func mustExec(t *testing.T, db *database.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
}

func count(t *testing.T, db *database.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(t.Context(), query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestMerge(t *testing.T) {
	db, q := mergeEnv(t)
	ctx := context.Background()
	loser := newItem(t, db, "game", "portal-dup", "Q1", "public")
	winner := newItem(t, db, "game", "portal", "Q2", "public")
	studio := newEntity(t, db, "organization", "valve")

	for _, c := range [][]any{{loser, "en", "Portal dup"}, {loser, "pt-BR", "Portal PT"}, {winner, "en", "Portal"}} {
		mustExec(
			t,
			db,
			`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1,$2,$3,'wikidata')`,
			c...)
	}
	mustExec(
		t,
		db,
		`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1,$2,'developer','wikidata'), ($3,$2,'developer','wikidata')`,
		loser,
		studio,
		winner,
	)
	mustExec(t, db, `UPDATE items SET external_ids = '{"igdb":"1","steam":"old"}' WHERE id = $1`, loser)
	mustExec(t, db, `UPDATE items SET external_ids = '{"steam":"new"}' WHERE id = $1`, winner)
	mustExec(
		t,
		db,
		`INSERT INTO users (id, name, email, password, created_at, updated_at) VALUES (1,'A','a@x.io','x',now(),now()), (2,'B','b@x.io','x',now(),now())`,
	)
	rate := `INSERT INTO ratings (user_id, item_id, c1,c2,c3,c4,c5,c6, updated_at) VALUES ($1,$2,$3,$3,$3,$3,$3,$3,$4::timestamptz)`
	mustExec(t, db, rate, 1, loser, 100, "2026-02-01")
	mustExec(t, db, rate, 1, winner, 50, "2026-01-01") // loser's is newer: wins
	mustExec(t, db, rate, 2, loser, 20, "2026-01-01")
	mustExec(t, db, rate, 2, winner, 80, "2026-02-01") // winner's is newer: stays

	if err := catalog.Merge(ctx, db, q, loser, winner, 0); err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1`, loser); n != 0 {
		t.Fatal("source not deleted")
	}
	if n := count(t, db, `SELECT count(*) FROM item_translations WHERE item_id = $1`, winner); n != 2 {
		t.Fatalf("translations = %d, want 2 (en kept, pt-BR copied)", n)
	}
	if n := count(
		t,
		db,
		`SELECT count(*) FROM item_translations WHERE item_id = $1 AND locale='en' AND title='Portal'`,
		winner,
	); n != 1 {
		t.Fatal("target translation must win")
	}
	if n := count(t, db, `SELECT count(*) FROM item_entities WHERE item_id = $1`, winner); n != 1 {
		t.Fatalf("entities = %d, want 1", n)
	}
	if n := count(
		t,
		db,
		`SELECT count(*) FROM items WHERE id = $1 AND external_ids = '{"igdb":"1","steam":"new"}'`,
		winner,
	); n != 1 {
		t.Fatal("external_ids not unioned with target winning")
	}
	if n := count(t, db, `SELECT c1 FROM ratings WHERE user_id = 1 AND item_id = $1`, winner); n != 100 {
		t.Fatalf("user 1 rating = %d, want the newer 100", n)
	}
	if n := count(t, db, `SELECT c1 FROM ratings WHERE user_id = 2 AND item_id = $1`, winner); n != 80 {
		t.Fatalf("user 2 rating = %d, want the newer 80", n)
	}
	got, ok, err := catalog.ResolveID(ctx, db, catalog.ScopeItem, "game", loser)
	if err != nil || !ok || got.ID != winner {
		t.Fatalf("id redirect = %+v %v %v", got, ok, err)
	}
	if got, ok, _ := catalog.ResolveSlug(ctx, db, catalog.ScopeItem, "game", "portal-dup"); !ok || got.ID != winner {
		t.Fatal("slug redirect missing")
	}
	if n := count(t, db, `SELECT count(*) FROM catalog_merge_audit
WHERE source_item_id = $1 AND target_item_id = $2
AND snapshot #>> '{source,item,slug}' = 'portal-dup' AND jsonb_array_length(snapshot #> '{source,translations}') = 2
AND snapshot #>> '{target,item,slug}' = 'portal'`, loser, winner); n != 1 {
		t.Fatal("snapshot row missing or incomplete")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM catalog_merge_audit`); err == nil {
		t.Fatal("audit table must be append-only")
	}
}

func TestMergeRefusals(t *testing.T) {
	db, q := mergeEnv(t)
	ctx := context.Background()
	a := newItem(t, db, "game", "a", "Q1", "public")
	b := newItem(t, db, "game", "b", "Q2", "public")
	m := newItem(t, db, "movie", "m", "Q3", "public")

	for name, c := range map[string]struct {
		from, into string
		want       error
	}{
		"self":     {a, a, catalog.ErrMergeSelf},
		"type":     {a, m, catalog.ErrMergeTypeMismatch},
		"missing":  {a, "00000000-0000-7000-8000-000000000000", catalog.ErrMergeNotFound},
		"missing2": {"00000000-0000-7000-8000-000000000000", a, catalog.ErrMergeNotFound},
	} {
		if err := catalog.Merge(ctx, db, q, c.from, c.into, 0); !errors.Is(err, c.want) {
			t.Errorf("%s: %v, want %v", name, err, c.want)
		}
	}

	if err := catalog.Merge(ctx, db, q, a, b, 0); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Merge(ctx, db, q, a, b, 0); !errors.Is(err, catalog.ErrMergeRedirected) {
		t.Fatalf("second merge: %v, want ErrMergeRedirected", err)
	}
}

func TestMergeRollsBackOnFailure(t *testing.T) {
	db, q := mergeEnv(t)
	ctx := context.Background()
	a := newItem(t, db, "game", "a", "Q1", "public")
	b := newItem(t, db, "game", "b", "Q2", "public")
	mustExec(
		t,
		db,
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1,'en','A','wikidata')`,
		a,
	)
	// Makes the final DELETE of the source fail after every move happened.
	mustExec(
		t,
		db,
		`CREATE FUNCTION block_delete() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'blocked'; END $$ LANGUAGE plpgsql;
CREATE TRIGGER block_delete BEFORE DELETE ON items FOR EACH ROW EXECUTE FUNCTION block_delete()`,
	)

	if err := catalog.Merge(ctx, db, q, a, b, 0); err == nil {
		t.Fatal("expected failure")
	}
	if n := count(t, db, `SELECT count(*) FROM item_translations WHERE item_id = $1`, a); n != 1 {
		t.Fatal("translation move not rolled back")
	}
	if n := count(
		t,
		db,
		`SELECT (SELECT count(*) FROM catalog_redirects) + (SELECT count(*) FROM catalog_merge_audit)`,
	); n != 0 {
		t.Fatal("redirect or audit row survived rollback")
	}
}
