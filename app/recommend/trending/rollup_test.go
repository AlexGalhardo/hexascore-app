//go:build integration

package trending_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

var now = time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

type rig struct {
	t   testing.TB
	db  *database.DB
	ctx context.Context
	n   int
}

func newRig(tb testing.TB) *rig {
	tb.Helper()

	url := pgtest.Database(tb, "")
	app := vukemtest.New(tb, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})

	db, err := app.DB()
	if err != nil {
		tb.Fatal(err)
	}

	return &rig{t: tb, db: db, ctx: context.Background()}
}

func (r *rig) exec(q string, args ...any) {
	r.t.Helper()

	if _, err := r.db.ExecContext(r.ctx, q, args...); err != nil {
		r.t.Fatalf("%v\n%s", err, q)
	}
}

func (r *rig) item() string {
	r.n++
	id := uuid.NewV7().String()
	r.exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, 'game', $2, $3, 'public')`,
		id, fmt.Sprintf("item-%d", r.n), fmt.Sprintf("Q%d", r.n))

	return id
}

// user creates an account aged ageDays at now; banned accounts carry banned_at.
func (r *rig) user(ageDays int, banned bool) int64 {
	r.n++

	var banAt *time.Time
	if banned {
		banAt = &now
	}

	var id int64

	err := r.db.QueryRowContext(r.ctx, `INSERT INTO users (name, email, password, created_at, banned_at)
VALUES ($1, $2, 'x', $3, $4) RETURNING id`, fmt.Sprintf("u%d", r.n), fmt.Sprintf("u%d@x.test", r.n),
		now.AddDate(0, 0, -ageDays), banAt).Scan(&id)
	if err != nil {
		r.t.Fatal(err)
	}

	return id
}

func (r *rig) event(kind string, user int64, item string, at time.Time, weight any) {
	r.exec(`INSERT INTO rating_events (at, kind, user_id, item_id, item_type, weight)
VALUES ($1, $2::rating_event_kind, $3, $4, 'game', $5)`, at, kind, user, item, weight)
}

func (r *rig) list(user int64, public, hidden bool, item string, addedAt time.Time) {
	id := uuid.NewV7().String()
	r.exec(`INSERT INTO lists (id, user_id, type, title, is_public, hidden) VALUES ($1, $2, 'game', 'l', $3, $4)`,
		id, user, public, hidden)
	r.exec(`INSERT INTO list_items (list_id, item_id, position, added_at) VALUES ($1, $2, 1, $3)`, id, item, addedAt)
}

func (r *rig) rollup() {
	r.t.Helper()

	if err := trending.Rollup(r.ctx, r.db, now); err != nil {
		r.t.Fatal(err)
	}
}

type counts struct {
	n int
	w float64
}

// row returns (count, weight) for one primitive of an item's day; the zero value when there is no row.
func (r *rig) row(item, day, col string) counts {
	r.t.Helper()

	var c counts

	var w float32

	q := fmt.Sprintf(`SELECT %s, %s_w FROM item_activity_daily WHERE item_id = $1 AND day = $2`, col, wcol(col))

	err := r.db.QueryRowContext(r.ctx, q, item, day).Scan(&c.n, &w)
	if err != nil && err.Error() != "sql: no rows in result set" {
		r.t.Fatal(err)
	}

	c.w = float64(w)

	return c
}

func wcol(col string) string {
	switch col {
	case "status_want":
		return "want"
	case "status_in_progress":
		return "in_progress"
	case "status_completed":
		return "completed"
	case "status_dropped":
		return "dropped"
	default:
		return col
	}
}

func (r *rig) want(item, day, col string, n int, w float64) {
	r.t.Helper()

	got := r.row(item, day, col)
	if got.n != n || math.Abs(got.w-w) > 1e-4 {
		r.t.Errorf("%s %s: got %d/%.3f, want %d/%.3f", day, col, got.n, got.w, n, w)
	}
}

const today, yesterday = "2026-10-31", "2026-10-30"

func TestRollupWeights(t *testing.T) {
	r := newRig(t)
	item := r.item()
	old, mid, fresh := r.user(400, false), r.user(40, false), r.user(3, false)
	banned, flagged := r.user(400, true), r.user(400, false)
	at := now.Add(-2 * time.Hour)

	r.event("rating_created", old, item, at, nil)
	r.event("rating_updated", old, item, at, nil)
	r.event("rating_created", mid, item, at, 0.5)
	r.event("rating_created", mid, item, at, 0.75)
	r.event("rating_created", fresh, item, at, nil)
	r.event("rating_created", banned, item, at, nil)
	r.event("rating_created", flagged, item, at, nil)
	r.exec(`INSERT INTO rating_flags (user_id, item_id, shadow_weight, reason)
VALUES ($1, $2, 0.5, 'admin_downweight')`, flagged, item)

	// created then deleted the same day: contributes nothing
	gone := r.user(400, false)
	r.event("rating_created", gone, item, at, nil)
	r.event("rating_deleted", gone, item, at.Add(time.Hour), nil)

	r.exec(`INSERT INTO item_status (user_id, item_id, status, updated_at)
VALUES ($1, $3, 'completed', $4), ($2, $3, 'want', $4)`, old, mid, item, at)

	sug := uuid.NewV7().String()
	r.exec(`INSERT INTO suggestions (id, author_id, kind, target_type, target_id, target_domain, locale,
  changes, status)
VALUES ($1, $2, 'edit', 'item', $3, 'game', 'en', '{}', 'submitted')`, sug, old, item)
	r.exec(`INSERT INTO suggestion_events (suggestion_id, to_status, actor_role, created_at)
VALUES ($1, 'draft', 'author', $2), ($1, 'submitted', 'author', $2)`, sug, at)

	r.rollup()

	r.want(item, today, "ratings_new", 4, 1+0.75+0.25+0.5)
	r.want(item, today, "ratings_updated", 1, 1)
	r.want(item, today, "status_completed", 1, 1)
	r.want(item, today, "status_want", 1, 0.75)
	r.want(item, today, "suggestions", 1, 1)
	r.want(item, yesterday, "ratings_new", 0, 0)
}

func TestListRules(t *testing.T) {
	r := newRig(t)
	item := r.item()
	old, fresh, banned := r.user(400, false), r.user(3, false), r.user(400, true)
	at := now.Add(-time.Hour)

	r.list(old, true, false, item, at)
	r.list(old, true, false, item, at) // same user twice: one count
	r.list(old, false, false, item, at)
	r.list(old, true, true, item, at)
	r.list(fresh, true, false, item, at)
	r.list(banned, true, false, item, at)
	r.list(r.user(40, false), true, false, item, at.Add(-24*time.Hour))

	r.rollup()

	r.want(item, today, "list_adds", 1, 1)
	r.want(item, yesterday, "list_adds", 1, 0.75)
}

func TestRollupIdempotent(t *testing.T) {
	r := newRig(t)
	item := r.item()
	old := r.user(400, false)
	r.event("rating_created", old, item, now.Add(-time.Hour), nil)
	r.event("rating_created", old, item, now.AddDate(0, 0, -3), nil)
	r.exec(`INSERT INTO item_activity_daily (item_id, day, item_type, views, ratings_new, ratings_new_w)
VALUES ($1, $2, 'game', 7, 9, 9), ($1, '2026-10-20', 'game', 7, 9, 9)`, item, yesterday)

	snapshot := func() string {
		var s string

		err := r.db.QueryRowContext(r.ctx, `SELECT coalesce(string_agg(t::text, ';' ORDER BY day), '')
FROM item_activity_daily t`).Scan(&s)
		if err != nil {
			t.Fatal(err)
		}

		return s
	}

	r.rollup()
	first := snapshot()
	r.rollup()

	if second := snapshot(); first != second {
		t.Fatalf("rollup is not idempotent:\n%s\n%s", first, second)
	}

	r.want(item, today, "ratings_new", 1, 1)
	r.want(item, yesterday, "ratings_new", 0, 0) // stale replaced
	r.want(item, "2026-10-28", "ratings_new", 0, 0)
	r.want(item, "2026-10-20", "ratings_new", 9, 9) // outside the two-day window: untouched

	var views int
	if err := r.db.QueryRowContext(r.ctx, `SELECT views FROM item_activity_daily WHERE item_id = $1 AND day = $2`,
		item, yesterday).Scan(&views); err != nil || views != 7 {
		t.Fatalf("views must survive the rollup: %d, %v", views, err)
	}
}
