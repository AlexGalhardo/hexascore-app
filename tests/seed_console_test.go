//go:build integration

package tests

import (
	"bytes"
	"context"
	"strings"
	"testing"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/console"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type seedFixture struct {
	app          *vukemtest.App
	db           *database.DB
	out          *bytes.Buffer
	fake         *queue.Fake
	sharedItem   string
	seedOnlyItem string
}

// newSeedFixture builds: two seed users (a, b), a real user on the seed email domain without a legacy id, a real
// user on the seed legacy prefix with another domain, and an admin matching both filters.
func newSeedFixture(t *testing.T) *seedFixture {
	t.Helper()

	f := &seedFixture{out: &bytes.Buffer{}, fake: queue.NewFake()}
	f.app = vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:        map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
		AppOptions: []vukem.Option{vukem.WithStdout(f.out), vukem.WithQueueDriver(f.fake)},
	})

	db, err := f.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	f.db = db
	f.sharedItem, f.seedOnlyItem = uuid.NewV7().String(), uuid.NewV7().String()
	f.exec(
		t,
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', 'seed-shared', 'Q1'), ($2, 'game', 'seed-only', 'Q2')`,
		f.sharedItem,
		f.seedOnlyItem,
	)

	user := func(email, legacy, role string) int64 {
		var id int64

		err := db.QueryRowContext(t.Context(), `INSERT INTO users (name, email, password, role, legacy_id, created_at, updated_at)
VALUES ('n', $1, 'x', $3, NULLIF($2, ''), now(), now()) RETURNING id`, email, legacy, role).
			Scan(&id)
		if err != nil {
			t.Fatal(err)
		}

		return id
	}

	seedA := user("a@example.test", "seed-user-1", "user")
	seedB := user("b@example.test", "seed-user-2", "user")
	realUser := user("real@example.test", "", "user")
	user("old@real.org", "seed-user-3", "user")
	user("root@example.test", "seed-user-4", "admin")

	rate := `INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 5, 5, 5, 5, 5, 5)`
	for _, u := range []int64{seedA, seedB, realUser} {
		f.exec(t, rate, u, f.sharedItem)
	}

	f.exec(t, rate, seedA, f.seedOnlyItem)
	f.exec(t, `INSERT INTO lists (id, user_id, type, title) VALUES (gen_random_uuid(), $1, 'game', 'l')`, seedA)
	f.exec(
		t,
		`INSERT INTO account_events (user_id, actor_id, kind, data, created_at) VALUES ($1, $2, 'x', '{}', now())`,
		realUser,
		seedA,
	)

	return f
}

func (f *seedFixture) exec(t *testing.T, q string, args ...any) {
	t.Helper()

	if _, err := f.db.ExecContext(t.Context(), q, args...); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

func (f *seedFixture) count(t *testing.T, q string, args ...any) int {
	t.Helper()

	var n int
	if err := f.db.QueryRowContext(t.Context(), q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func (f *seedFixture) run(ctx context.Context, name string, args ...string) error {
	f.out.Reset()

	return f.app.Call(ctx, name, args)
}

func TestSeedMark(t *testing.T) {
	f := newSeedFixture(t)
	args := []string{"--email-suffix", "@example.test", "--legacy-prefix", "seed-user-"}

	if err := f.run(t.Context(), "seed:mark", append(args, "--dry-run")...); err != nil {
		t.Fatal(err)
	}

	if n := f.count(t, `SELECT count(*) FROM users WHERE is_seed`); n != 0 {
		t.Fatalf("dry run marked %d users", n)
	}

	if !strings.Contains(f.out.String(), "Matching users: 2") {
		t.Fatalf("dry-run output: %s", f.out.String())
	}

	if err := f.run(t.Context(), "seed:mark", args...); err != nil {
		t.Fatal(err)
	}

	if n := f.count(
		t,
		`SELECT count(*) FROM users WHERE is_seed AND email IN ('a@example.test','b@example.test')`,
	); n != 2 {
		t.Fatalf("seed users marked = %d, want 2", n)
	}

	// The real-looking users and the admin that matches both filters stay unmarked.
	if n := f.count(t, `SELECT count(*) FROM users WHERE is_seed`); n != 2 {
		t.Fatalf("marked users = %d, want 2", n)
	}

	if err := f.run(t.Context(), "seed:mark"); err == nil {
		t.Fatal("seed:mark without filters succeeded")
	}
}

func TestSeedPurge(t *testing.T) {
	f := newSeedFixture(t)
	f.exec(t, `UPDATE users SET is_seed = true WHERE email IN ('a@example.test','b@example.test')`)

	if err := f.run(t.Context(), "seed:purge"); err == nil {
		t.Fatal("seed:purge ran without --confirm")
	}

	if err := f.run(t.Context(), "seed:purge", "--dry-run"); err != nil {
		t.Fatal(err)
	}

	if n := f.count(t, `SELECT count(*) FROM users`); n != 5 {
		t.Fatalf("dry run deleted users, %d left", n)
	}

	if !strings.Contains(f.out.String(), "ratings.user_id") || !strings.Contains(f.out.String(), "Seed users: 2") {
		t.Fatalf("dry-run output: %s", f.out.String())
	}

	if err := f.run(t.Context(), "seed:purge", "--confirm", "--batch", "1"); err != nil {
		t.Fatal(err)
	}

	if n := f.count(t, `SELECT count(*) FROM users WHERE is_seed`); n != 0 {
		t.Fatalf("seed users left: %d", n)
	}

	realLeft := `SELECT count(*) FROM users WHERE email IN ('real@example.test','old@real.org','root@example.test')`
	if n := f.count(t, realLeft); n != 3 {
		t.Fatalf("real users left = %d, want 3", n)
	}

	// Seed ratings and lists are gone, the real user's rating stays, SET NULL keeps the event.
	if n := f.count(t, `SELECT count(*) FROM ratings`); n != 1 {
		t.Fatalf("ratings left = %d, want 1", n)
	}

	if n := f.count(t, `SELECT count(*) FROM lists`); n != 0 {
		t.Fatalf("lists left = %d", n)
	}

	if n := f.count(t, `SELECT count(*) FROM account_events WHERE actor_id IS NULL AND kind = 'x'`); n != 1 {
		t.Fatal("SET NULL event not kept")
	}

	dirty := f.count(t, `SELECT count(*) FROM search_dirty WHERE kind = 'item' AND ref_id IN ($1::uuid, $2::uuid)`,
		f.sharedItem, f.seedOnlyItem)
	if dirty != 2 {
		t.Fatalf("search_dirty items = %d, want 2", dirty)
	}

	jobs := 0

	for _, m := range f.fake.Messages() {
		if m.Kind == "scoring.recompute_item_score" {
			jobs++
		}
	}

	if jobs != 2 {
		t.Fatalf("recompute jobs = %d, want 2", jobs)
	}
}

// Every column that references users.id must be listed in console.SeedTables, or purge would leave a row
// behind (loose columns are not FKs) or fail on a RESTRICT foreign key.
func TestSeedPurgeCoversEveryUserFK(t *testing.T) {
	f := newSeedFixture(t)

	rows, err := f.db.QueryContext(t.Context(), `SELECT c.conrelid::regclass::text, a.attname, c.confdeltype::text
FROM pg_constraint c JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)
WHERE c.contype = 'f' AND c.confrelid = 'users'::regclass`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	listed := map[string]string{}

	for _, s := range console.SeedTables {
		if s.Kind != "loose" {
			listed[s.Table+"."+s.Column] = s.Kind
		}
	}

	seen := 0

	for rows.Next() {
		var table, col, del string
		if err := rows.Scan(&table, &col, &del); err != nil {
			t.Fatal(err)
		}

		want := map[string]string{"c": "cascade", "n": "setnull"}[del]
		if listed[table+"."+col] != want {
			t.Errorf("%s.%s (confdeltype %q) listed as %q", table, col, del, listed[table+"."+col])
		}

		seen++
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if seen != len(listed) {
		t.Errorf("schema has %d user FKs, SeedTables lists %d", seen, len(listed))
	}
}
