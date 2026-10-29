//go:build integration

package scoring_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand" //nolint:depguard // reproducible fixtures, not secrets
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

const aggregateCols = `rating_count, rating_weight, score, score_c1, score_c2, score_c3, score_c4, score_c5, score_c6,
score_se, score_confidence, score_formula_version`

// aggregate renders the stored columns of an item; score_computed_at is left out because it is the compute time.
func aggregate(ctx context.Context, q *sql.DB, itemID string) (string, error) {
	var s string
	err := q.QueryRowContext(ctx, `SELECT row(`+aggregateCols+`)::text FROM items WHERE id = $1`, itemID).Scan(&s)

	return s, err
}

// nightly is the nightly job's per-item step (SDD 02 §7.2): lock the item, then recompute.
func nightly(ctx context.Context, db *database.DB, itemID string) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		var typ string
		if err := tx.QueryRowContext(ctx, `SELECT type FROM items WHERE id = $1 FOR UPDATE`, itemID).
			Scan(&typ); err != nil {
			return err
		}
		p, err := activeParams(ctx, tx)
		if err != nil {
			return err
		}

		return scoring.Recompute(ctx, tx, itemID, time.Now().UTC(), p)
	})
}

// activeParams reads formula version and game m_t; priors stay at the default 60 (no score_priors rows in these tests).
func activeParams(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
},
) (scoring.Params, error) {
	p := scoring.Params{Prior: [6]float64{60, 60, 60, 60, 60, 60}}
	err := q.QueryRowContext(ctx, `SELECT version, m_game FROM score_formula_versions
WHERE activated_at <= now() ORDER BY version DESC LIMIT 1`).Scan(&p.Version, &p.M)

	return p, err
}

// TestConcurrentWriters runs 50 writers on one item plus repeated nightly recomputes, then requires the
// stored aggregate to equal a from-scratch recompute and the rating count to equal the row count (NFR-02-4).
func TestConcurrentWriters(t *testing.T) {
	const writers, ops = 50, 3

	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx, q := t.Context(), db.SQL()
	r := &rig{t: t, app: app}
	item := r.item("game", "2020-01-01", "day", "released")

	users := make([]int64, writers)
	for i := range users {
		if err := q.QueryRowContext(ctx, `INSERT INTO users (name, email, password, created_at, updated_at)
VALUES ('U', $1, 'x', now() - interval '200 days', now()) RETURNING id`, fmt.Sprintf("cw%d@example.com", i)).Scan(&users[i]); err != nil {
			t.Fatal(err)
		}
		// A non-"want" status lets the user through the burst limiter's item gate.
		if _, err := q.ExecContext(
			ctx,
			`INSERT INTO item_status (user_id, item_id, status) VALUES ($1, $2, 'completed')`,
			users[i],
			item,
		); err != nil {
			t.Fatal(err)
		}
	}

	w := &scoring.Writer{DB: db, Now: time.Now}
	keys := scoring.CategoryKeys["game"]
	errc := make(chan error, writers*ops+16)
	var wg sync.WaitGroup

	for i, uid := range users {
		wg.Go(func() {
			rng := rand.New(rand.NewSource(int64(i)))
			for range ops {
				var err error
				if rng.Intn(10) < 3 {
					_, err = w.Delete(ctx, uid, item)
				} else {
					stars := map[string]float64{}
					for _, k := range keys {
						stars[k] = float64(1+rng.Intn(10)) / 2
					}
					_, err = w.Save(ctx, uid, item, stars)
				}
				if err != nil {
					errc <- fmt.Errorf("writer %d: %w", i, err)
				}
			}
		})
	}
	wg.Go(func() {
		for range 8 {
			if err := nightly(ctx, db, item); err != nil {
				errc <- fmt.Errorf("nightly: %w", err)
			}
		}
	})
	wg.Wait()
	close(errc)
	for err := range errc {
		t.Error(err)
	}

	got, err := aggregate(ctx, q, item)
	if err != nil {
		t.Fatal(err)
	}
	if err := nightly(ctx, db, item); err != nil {
		t.Fatal(err)
	}
	want, err := aggregate(ctx, q, item)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("stored aggregate diverged from a from-scratch recompute:\n got  %s\n want %s", got, want)
	}

	var rows, count int
	if err := q.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM ratings WHERE item_id = $1), rating_count FROM items WHERE id = $1`, item).
		Scan(&rows, &count); err != nil {
		t.Fatal(err)
	}
	if rows != count {
		t.Errorf("rating_count %d, ratings rows %d", count, rows)
	}
}

// TestRecomputeLatencyHotItem records recompute p50/p95 on hot items with 5,000 and 10,000 ratings (open
// issue 02-3, NFR-02-1) next to 40 small items, so the planner sees a realistic table.
func TestRecomputeLatencyHotItem(t *testing.T) {
	const runs = 100

	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx, q := t.Context(), db.SQL()
	r := &rig{t: t, app: app}

	if _, err := q.ExecContext(ctx, `INSERT INTO users (name, email, password, created_at, updated_at)
SELECT 'U', 'hot' || g || '@example.com', 'x', now() - (g % 200) * interval '1 day', now() FROM generate_series(1, 10000) g`); err != nil {
		t.Fatal(err)
	}
	fill := func(item string, n int) {
		for _, stmt := range []string{
			`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at, updated_at)
SELECT id, $1, 5 * (1 + id % 20), 5 * (1 + (id * 3) % 20), 5 * (1 + (id * 7) % 20),
  5 * (1 + (id * 11) % 20), 5 * (1 + (id * 13) % 20), 5 * (1 + (id * 17) % 20), now(), now() FROM users ORDER BY id LIMIT $2`,
			`INSERT INTO item_status (user_id, item_id, status) SELECT user_id, item_id, 'completed' FROM ratings
WHERE item_id = $1 AND user_id % 2 = 0 AND $2 > 0`,
		} {
			if _, err := q.ExecContext(ctx, stmt, item, n); err != nil {
				t.Fatal(err)
			}
		}
	}
	for range 40 {
		fill(r.item("game", "2020-01-01", "day", "released"), 100)
	}
	p, err := activeParams(ctx, q)
	if err != nil {
		t.Fatal(err)
	}

	for _, n := range []int{5000, 10000} {
		item := r.item("game", "2020-01-01", "day", "released")
		fill(item, n)
		if _, err := q.ExecContext(ctx, `ANALYZE`); err != nil {
			t.Fatal(err)
		}

		lat := make([]time.Duration, runs)
		for i := range lat {
			start := time.Now()
			if err := scoring.Recompute(ctx, q, item, time.Now().UTC(), p); err != nil {
				t.Fatal(err)
			}
			lat[i] = time.Since(start)
		}
		slices.Sort(lat)
		t.Logf(
			"recompute of %d ratings, %d runs: p50 %s p95 %s max %s",
			n,
			runs,
			lat[runs/2],
			lat[runs*95/100],
			lat[runs-1],
		)

		var count int
		if err := q.QueryRowContext(ctx, `SELECT rating_count FROM items WHERE id = $1`, item).
			Scan(&count); err != nil ||
			count != n {
			t.Fatalf("rating_count %d (err %v), want %d", count, err, n)
		}
	}
}
