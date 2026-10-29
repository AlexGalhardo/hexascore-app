//go:build integration

package scoring_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

func (r *rig) exec(q string, args ...any) {
	r.t.Helper()
	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}
	if _, err := db.SQL().ExecContext(r.t.Context(), q, args...); err != nil {
		r.t.Fatalf("%s: %v", q, err)
	}
}

func (r *rig) userID(n int) int64 {
	r.t.Helper()
	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}
	var id int64
	if err := db.SQL().
		QueryRowContext(r.t.Context(), `SELECT id FROM users WHERE email = $1`, fmt.Sprintf("u%d@example.com", n)).
		Scan(&id); err != nil {
		r.t.Fatal(err)
	}
	return id
}

// rateAs has user n (created on first use) rate item with every category at stars.
func (r *rig) rateAs(n int, item string, stars float64) {
	r.t.Helper()
	body := map[string]any{"scores": map[string]float64{
		"gameplay": stars, "fun": stars, "sound": stars, "technical": stars, "visual": stars, "storytelling": stars,
	}}
	r.token(n, "ratings:write").PutJSON("/app/v1/items/"+item+"/rating", body).AssertOK()
}

func TestBanRecompute(t *testing.T) {
	r := newRig(t)
	item := r.item("game", "2020-01-01", "day", "released")
	r.rateAs(1, item, 10)
	for i := 2; i <= 5; i++ {
		r.rateAs(i, item, 5)
	}
	r.token(9, "ratings:write")
	adminID, user1 := r.userID(9), r.userID(1)
	before := r.score(item)

	if err := r.acc.Ban(t.Context(), adminID, user1, "spam", nil); err != nil {
		t.Fatal(err)
	}
	if n := r.countOf(`SELECT rating_count FROM items WHERE id = $1`, item); n != 4 {
		t.Fatalf("rating_count after ban = %d, want 4", n)
	}
	if after := r.score(item); after >= before {
		t.Fatalf("score after ban = %v, want below %v", after, before)
	}

	if err := r.acc.Unban(t.Context(), adminID, user1); err != nil {
		t.Fatal(err)
	}
	if n := r.countOf(`SELECT rating_count FROM items WHERE id = $1`, item); n != 5 {
		t.Fatalf("rating_count after unban = %d, want 5", n)
	}
	if got := r.score(item); got != before {
		t.Fatalf("score after unban = %v, want %v", got, before)
	}
}

func TestMergeMovesRatings(t *testing.T) {
	r := newRig(t)
	loser := r.item("game", "2020-01-01", "day", "released")
	winner := r.item("game", "2020-01-01", "day", "released")
	r.exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`,
	)
	r.exec(
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1,'en','A','wikidata'), ($2,'en','B','wikidata')`,
		loser,
		winner,
	)
	for i := 1; i <= 3; i++ {
		r.rateAs(i, loser, 10)
	}
	for i := 4; i <= 5; i++ {
		r.rateAs(i, winner, 5)
	}

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	q, err := r.app.Queue()
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.Merge(t.Context(), db, q, loser, winner, 0); err != nil {
		t.Fatal(err)
	}

	if n := r.countOf(`SELECT rating_count FROM items WHERE id = $1`, winner); n != 5 {
		t.Fatalf("winner rating_count = %d, want 5", n)
	}
	if s := r.score(winner); s <= 60 {
		t.Fatalf("winner score = %v, want above the prior 60", s)
	}
}

func TestNightlyRepairsDrift(t *testing.T) {
	r := newRig(t)
	item := r.item("game", "2020-01-01", "day", "released")
	for i := 1; i <= 6; i++ {
		r.rateAs(i, item, 8)
	}
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	want := r.score(item)

	r.exec(
		`UPDATE items SET rating_count = 2, score = 3, score_computed_at = now() - interval '2 days' WHERE id = $1`,
		item,
	)
	drift := func() int {
		rep, err := catalog.RunDQ(t.Context(), db, scoring.DQAsserts()...)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, c := range rep.Checks {
			if len(c.ID) > 2 && c.ID[:2] == "SC" {
				n += c.Count
			}
		}
		return n
	}
	if drift() == 0 {
		t.Fatal("DQ missed the tampered item")
	}

	now := time.Now().UTC()
	if err := scoring.RebuildPriors(t.Context(), db, now); err != nil {
		t.Fatal(err)
	}
	if _, err := scoring.RecomputeAll(t.Context(), db, now); err != nil {
		t.Fatal(err)
	}
	if n := drift(); n != 0 {
		t.Fatalf("drift after nightly = %d, want 0", n)
	}
	if got := r.score(item); got != want {
		t.Fatalf("score after nightly = %v, want %v", got, want)
	}
	if n := r.countOf(
		`SELECT count(*) FROM score_priors WHERE item_type = 'game' AND prior_c[1] = 60 AND items_used = 1`,
	); n != 1 {
		t.Fatal("game prior must stay 60 with fewer than 20 qualifying items")
	}
}

func TestReleaseBursts(t *testing.T) {
	r := newRig(t)
	item := r.item("game", "2020-01-01", "day", "released")
	r.rateAs(1, item, 8)
	var old, fresh int64
	db, _ := r.app.DB()
	ins := `INSERT INTO score_bursts (item_id, started_at, limited_until, trigger, stats)
VALUES ($1, now() - $2 * interval '1 day', now(), 'volume', '{}') RETURNING id`
	if err := db.SQL().QueryRowContext(t.Context(), ins, item, 15).Scan(&old); err != nil {
		t.Fatal(err)
	}
	if err := db.SQL().QueryRowContext(t.Context(), ins, item, 2).Scan(&fresh); err != nil {
		t.Fatal(err)
	}
	r.exec(
		`INSERT INTO rating_flags (user_id, item_id, shadow_weight, reason, burst_id) VALUES ($1, $2, 0.25, 'burst_hold', $3)`,
		r.userID(1),
		item,
		old,
	)

	n, err := scoring.ReleaseBursts(t.Context(), db, time.Now().UTC())
	if err != nil || n != 1 {
		t.Fatalf("resolved = %d err %v, want 1", n, err)
	}
	if c := r.countOf(`SELECT count(*) FROM score_bursts WHERE id = $1 AND resolution = 'auto_released'`, old); c != 1 {
		t.Fatal("old burst not auto_released")
	}
	if c := r.countOf(`SELECT count(*) FROM score_bursts WHERE id = $1 AND resolution IS NULL`, fresh); c != 1 {
		t.Fatal("fresh burst must stay open")
	}
	if c := r.countOf(`SELECT count(*) FROM rating_flags`); c != 0 {
		t.Fatal("flags of the released burst must be removed")
	}
	if c := r.countOf(
		`SELECT count(*) FROM score_actions WHERE kind = 'burst_auto_release' AND actor IS NULL`,
	); c != 1 {
		t.Fatal("system audit row missing")
	}
}

func TestPruneRatingEvents(t *testing.T) {
	r := newRig(t)
	item := r.item("game", "2020-01-01", "day", "released")
	r.exec(`INSERT INTO rating_events (at, kind, user_id, item_id, item_type)
VALUES (now() - interval '181 days', 'rating_created', 1, $1, 'game'), (now() - interval '179 days', 'rating_created', 2, $1, 'game')`, item)
	db, _ := r.app.DB()
	n, err := scoring.PruneRatingEvents(t.Context(), db, time.Now().UTC())
	if err != nil || n != 1 {
		t.Fatalf("pruned = %d err %v, want 1", n, err)
	}
	if c := r.countOf(`SELECT count(*) FROM rating_events`); c != 1 {
		t.Fatalf("events left = %d, want 1", c)
	}
}
