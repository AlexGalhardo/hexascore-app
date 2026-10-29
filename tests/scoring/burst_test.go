//go:build integration

package scoring_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

func (r *rig) countOf(q string, args ...any) int {
	r.t.Helper()
	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}
	var n int
	if err := db.SQL().QueryRowContext(r.t.Context(), q, args...).Scan(&n); err != nil {
		r.t.Fatal(err)
	}
	return n
}

func (r *rig) score(item string) float64 {
	r.t.Helper()
	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}
	var s float64
	if err := db.SQL().
		QueryRowContext(r.t.Context(), `SELECT score FROM items WHERE id = $1`, item).
		Scan(&s); err != nil {
		r.t.Fatal(err)
	}
	return s
}

// TestBurst seeds a baseline b = 2 and 40 young-account inserts in one hour:
// one burst opens at the 15th insert (1 h rule: max(15, 3b)), the rest are limited, the window is held
// at 0.25 and an admin approve restores the unflagged aggregate exactly.
func TestBurst(t *testing.T) {
	r := newRig(t)
	item := r.item("game", "2020-01-01", "day", "released")
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().ExecContext(t.Context(), `INSERT INTO rating_events (at, kind, user_id, item_id, item_type)
SELECT now() - interval '5 days', 'rating_created', 900000 + g, $1, 'game' FROM generate_series(1, 56) g`, item); err != nil {
		t.Fatal(err)
	}

	path := "/app/v1/items/" + item + "/rating"
	for i := 1; i <= 40; i++ {
		resp := r.tokenAged(i, 3, "ratings:write").PutJSON(path, gameScores)
		if i <= 15 {
			resp.AssertOK()
			continue
		}
		resp.AssertStatus(429)
		if resp.Header("Retry-After") == "" {
			t.Fatal("limited response has no Retry-After")
		}
	}

	if got := r.countOf(`SELECT count(*) FROM score_bursts WHERE item_id = $1 AND resolution IS NULL`, item); got != 1 {
		t.Fatalf("open bursts = %d, want 1", got)
	}
	if got := r.countOf(
		`SELECT count(*) FROM rating_flags WHERE item_id = $1 AND shadow_weight = 0.25 AND reason = 'burst_hold'`,
		item,
	); got != 15 {
		t.Fatalf("held ratings = %d, want 15", got)
	}
	if got := r.countOf(
		`SELECT count(*) FROM score_actions WHERE kind = 'burst_open' AND after IS NOT NULL`,
	); got != 1 {
		t.Fatalf("burst_open actions = %d, want 1", got)
	}

	var burstID int64
	if err := db.SQL().
		QueryRowContext(t.Context(), `SELECT id FROM score_bursts WHERE item_id = $1`, item).
		Scan(&burstID); err != nil {
		t.Fatal(err)
	}
	w := &scoring.Writer{DB: db, Now: time.Now}
	if err := w.ResolveBurst(t.Context(), 1, burstID, true); err != nil {
		t.Fatal(err)
	}
	if err := w.ResolveBurst(t.Context(), 1, burstID, true); err == nil {
		t.Fatal("second resolve succeeded")
	}
	if got := r.countOf(`SELECT count(*) FROM rating_flags WHERE item_id = $1`, item); got != 0 {
		t.Fatalf("flags left = %d", got)
	}
	got := r.score(item)
	if _, err := db.SQL().ExecContext(t.Context(), `UPDATE items SET score = NULL WHERE id = $1`, item); err != nil {
		t.Fatal(err)
	}
	// Equals a from-scratch recompute with no flags at all.
	if err := scoring.Recompute(t.Context(), db.SQL(), item, time.Now(), scoring.Params{
		M: 10, Version: 1,
		Prior: [scoring.Categories]float64{60, 60, 60, 60, 60, 60},
	}); err != nil {
		t.Fatal(err)
	}
	if want := r.score(item); got != want {
		t.Fatalf("score after approve = %v, want %v", got, want)
	}
	if got := r.countOf(
		`SELECT count(*) FROM items WHERE id = $1 AND ratings_limited_until IS NULL AND score_formula_version = 1`,
		item,
	); got != 1 {
		t.Fatal("limits not lifted or formula version not stored")
	}
}

// TestDailyCap: an account under 7 days old may create 20 ratings per UTC day.
func TestDailyCap(t *testing.T) {
	r := newRig(t)
	c := r.tokenAged(1, 2, "ratings:write")
	for i := range 21 {
		item := r.item("game", "2020-01-01", "day", "released")
		resp := c.PutJSON("/app/v1/items/"+item+"/rating", gameScores)
		if i < 20 {
			resp.AssertOK()
			continue
		}
		resp.AssertStatus(429)
		if resp.Header("Retry-After") == "" {
			t.Fatal("no Retry-After")
		}
	}
}

// TestSecurity_A04_ExcludeBurst: exclude holds the window at weight 0 and the
// moderation routes are closed to non-admins.
func TestSecurity_A04_ExcludeBurst(t *testing.T) {
	r := newRig(t)
	item := r.item("game", "2020-01-01", "day", "released")
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().ExecContext(t.Context(), `INSERT INTO rating_events (at, kind, user_id, item_id, item_type)
SELECT now() - interval '5 days', 'rating_created', 900000 + g, $1, 'game' FROM generate_series(1, 56) g`, item); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 15; i++ {
		r.tokenAged(i, 3, "ratings:write").PutJSON("/app/v1/items/"+item+"/rating", gameScores).AssertOK()
	}
	var burstID int64
	if err := db.SQL().
		QueryRowContext(t.Context(), `SELECT id FROM score_bursts WHERE item_id = $1`, item).
		Scan(&burstID); err != nil {
		t.Fatal(err)
	}

	r.app.Client(t).WithHeader("Sec-Fetch-Site", "same-origin").
		PostJSON(fmt.Sprintf("/admin/scoring/bursts/%d/exclude", burstID), map[string]any{}).AssertStatus(401)
	var userID int64
	if err := db.SQL().
		QueryRowContext(t.Context(), `INSERT INTO users (name, email, password, email_verified_at, created_at, updated_at)
VALUES ('N', 'n@example.com', 'x', now(), now(), now()) RETURNING id`).
		Scan(&userID); err != nil {
		t.Fatal(err)
	}
	r.app.Client(t).WithHeader("Sec-Fetch-Site", "same-origin").ActingAs("web", authID(userID)).
		PostJSON(fmt.Sprintf("/admin/scoring/bursts/%d/exclude", burstID), map[string]any{}).AssertStatus(404)
	if got := r.countOf(
		`SELECT count(*) FROM rating_flags WHERE item_id = $1 AND shadow_weight = 0.25`,
		item,
	); got != 15 {
		t.Fatalf("unauthorised requests changed flags: %d", got)
	}

	w := &scoring.Writer{DB: db, Now: time.Now}
	if err := w.ResolveBurst(t.Context(), 1, burstID, false); err != nil {
		t.Fatal(err)
	}
	if got := r.countOf(
		`SELECT count(*) FROM rating_flags WHERE item_id = $1 AND shadow_weight = 0 AND reason = 'admin_exclude'`,
		item,
	); got != 15 {
		t.Fatalf("excluded flags = %d, want 15", got)
	}
	if got := r.countOf(`SELECT rating_count FROM items WHERE id = $1`, item); got != 0 {
		t.Fatalf("rating_count after exclude = %d, want 0 (all 15 weigh 0)", got)
	}
	vukemtest.AssertDatabaseHas(t, r.app, "score_actions", map[string]any{"kind": "burst_exclude"})
}
