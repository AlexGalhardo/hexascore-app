//go:build integration

package similar_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

var now = time.Date(2026, 10, 31, 2, 0, 0, 0, time.UTC)

type rig struct {
	t   testing.TB
	db  *database.DB
	ctx context.Context
	n   int
}

// newRig seeds 60 filler games so that a feature shared by two items stays under the 5 % candidate cut.
func newRig(tb testing.TB) *rig {
	tb.Helper()

	app := vukemtest.New(tb, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(tb, "")}})

	db, err := app.DB()
	if err != nil {
		tb.Fatal(err)
	}

	r := &rig{t: tb, db: db, ctx: context.Background()}
	r.exec(`INSERT INTO data_sources (code, name, licence, url) VALUES
('hexascore', 'x', 'proprietary', 'https://x'), ('other', 'y', 'proprietary', 'https://y')`)

	for range 60 {
		r.item("public")
	}

	return r
}

func (r *rig) exec(q string, args ...any) {
	r.t.Helper()

	if _, err := r.db.ExecContext(r.ctx, q, args...); err != nil {
		r.t.Fatalf("%v\n%s", err, q)
	}
}

func (r *rig) item(visibility string) string {
	r.n++
	id := uuid.NewV7().String()
	r.exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, 'game', $2, $3, $4)`,
		id, fmt.Sprintf("item-%d", r.n), fmt.Sprintf("Q%d", r.n), visibility)

	return id
}

func (r *rig) entity(kind string) string {
	r.n++
	id := uuid.NewV7().String()
	r.exec(`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, $2, $3, 'e')`,
		id, kind, fmt.Sprintf("e-%d", r.n))

	return id
}

func (r *rig) link(item, entity, role, source string) {
	r.exec(`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, $3, $4)`,
		item, entity, role, source)
}

func (r *rig) neighbours(item string) (ids string, scores string, ok bool) {
	r.t.Helper()

	err := r.db.QueryRowContext(r.ctx,
		`SELECT array_to_string(similar_ids, ','), array_to_string(scores, ',') FROM item_similar WHERE item_id = $1`,
		item).Scan(&ids, &scores)

	return ids, scores, err == nil
}

func (r *rig) run(opt similar.Options) int {
	r.t.Helper()

	opt.Pause = -1
	opt.Now = now

	n, err := similar.Compute(r.ctx, r.db, opt)
	if err != nil {
		r.t.Fatal(err)
	}

	return n
}

// pair builds two games sharing franchise and developer and a third sharing only a genre with the first.
func (r *rig) pair() (a, b, c string) {
	franchise, dev, genre := r.entity("franchise"), r.entity("organization"), r.entity("genre")
	a, b, c = r.item("public"), r.item("public"), r.item("public")

	for _, it := range []string{a, b} {
		r.link(it, franchise, "franchise", "hexascore")
		r.link(it, dev, "developer", "other")
	}

	r.link(a, genre, "genre", "hexascore")
	r.link(c, genre, "genre", "hexascore")

	return a, b, c
}

func TestSimilarJob(t *testing.T) {
	r := newRig(t)
	a, b, c := r.pair()

	hidden := r.item("hidden")
	r.link(hidden, r.entity("franchise"), "franchise", "hexascore")

	// A shared hidden-item feature must not leak: give hidden the same franchise as a.
	var fr string
	if err := r.db.QueryRowContext(r.ctx, `SELECT entity_id::text FROM item_entities WHERE item_id = $1 AND role = 'franchise'`, a).
		Scan(&fr); err != nil {
		t.Fatal(err)
	}

	r.exec(`DELETE FROM item_entities WHERE item_id = $1`, hidden)
	r.link(hidden, fr, "franchise", "hexascore")

	// two public lists put a and b together, one hidden does not count.
	for i := range 3 {
		list := uuid.NewV7().String()
		r.exec(`INSERT INTO users (name, email, password) VALUES ('u', $1, 'x')`, fmt.Sprintf("u%d@x.test", i))
		r.exec(`INSERT INTO lists (id, user_id, type, title, is_public, hidden, item_count)
VALUES ($1, (SELECT max(id) FROM users), 'game', 'l', true, $2, 2)`, list, i == 2)
		r.exec(`INSERT INTO list_items (list_id, item_id, position) VALUES ($1, $2, 1), ($1, $3, 2)`, list, a, b)
	}

	if n := r.run(similar.Options{}); n < 3 {
		t.Fatalf("written = %d, want at least the 3 featured items", n)
	}

	ids, scores, ok := r.neighbours(a)
	if !ok {
		t.Fatal("no row for a")
	}

	if ids[:36] != b {
		t.Fatalf("a: first neighbour %s, want b %s (all: %s)", ids, b, ids)
	}

	if ids != b && ids != b+","+c {
		t.Fatalf("a neighbours = %s, want b then maybe c, never hidden or itself", ids)
	}

	var co string
	if err := r.db.QueryRowContext(r.ctx, `SELECT array_to_string(cooccur, ',') FROM item_similar WHERE item_id = $1`, a).
		Scan(&co); err != nil {
		t.Fatal(err)
	}

	if co[:1] != "2" {
		t.Fatalf("cooccur = %s, want 2 for b (hidden list excluded)", co)
	}

	t.Logf("a -> %s scores %s", ids, scores)

	if _, _, ok := r.neighbours(hidden); ok {
		t.Fatal("hidden item must not get a row")
	}
}

func TestSimilarSkipsUnchanged(t *testing.T) {
	r := newRig(t)
	a, _, _ := r.pair()

	if r.run(similar.Options{}) == 0 {
		t.Fatal("first run wrote nothing")
	}

	if n := r.run(similar.Options{}); n != 0 {
		t.Fatalf("second run wrote %d rows, want 0 (feature_hash unchanged)", n)
	}

	r.link(a, r.entity("genre"), "genre", "hexascore")

	if n := r.run(similar.Options{}); n != 1 {
		t.Fatalf("after a feature change wrote %d rows, want 1", n)
	}
}

func TestApiVariantUsesAllowlistedSources(t *testing.T) {
	r := newRig(t)
	a, b, _ := r.pair()

	// b shares franchise (hexascore) and developer (other). With only 'hexascore' allowlisted, the developer drops out
	// of the score but the franchise keeps b a neighbour; with only 'other' the franchise drops out.
	_ = r.run(similar.Options{})
	_, siteScores, _ := r.neighbours(a)

	r.exec(`DELETE FROM item_similar`)
	_ = r.run(similar.Options{Sources: []string{"hexascore"}})
	ids, apiScores, ok := r.neighbours(a)

	if !ok || ids == "" || ids[:36] != b {
		t.Fatalf("api variant lost b: %q", ids)
	}

	if apiScores == siteScores {
		t.Fatalf(
			"api scores %s equal site scores %s: the developer from source 'other' was not filtered",
			apiScores,
			siteScores,
		)
	}

	r.exec(`DELETE FROM item_similar`)
	_ = r.run(similar.Options{Sources: []string{"nothing"}})

	if ids, _, _ := r.neighbours(a); ids != "" {
		t.Fatalf("no allowlisted source must give no neighbours, got %s", ids)
	}
}
