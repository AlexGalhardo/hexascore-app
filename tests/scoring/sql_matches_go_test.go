//go:build integration

package scoring_test

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand" //nolint:depguard // reproducible fixtures, not secrets
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestSQLMatchesGo recomputes random items in SQL and in Go and requires the
// same aggregate to 1e-6 (SDD 02 §10).
func TestSQLMatchesGo(t *testing.T) {
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx, q := t.Context(), db.SQL()
	rng := rand.New(rand.NewSource(7))
	at := time.Date(2026, 10, 29, 15, 0, 0, 0, time.UTC)
	p := scoring.Params{M: 10, Prior: [6]float64{60, 62.5, 64, 58, 70, 66}}
	statuses := []string{"", "want", "in_progress", "completed", "dropped"}
	precisions := []string{"day", "month", "year", "unknown"}

	type user struct {
		id      int64
		created time.Time
		banned  bool
	}
	var users []user
	for i := range 40 {
		u := user{created: at.Add(-time.Duration(rng.Intn(200*24)) * time.Hour), banned: rng.Intn(15) == 0}
		var banned any
		if u.banned {
			banned = at
		}
		if err := q.QueryRowContext(ctx, `INSERT INTO users (name, email, password, banned_at, created_at, updated_at)
VALUES ($1, $2, 'x', $3, $4, $4) RETURNING id`, fmt.Sprintf("U%d", i), fmt.Sprintf("u%d@example.com", i), banned, u.created).
			Scan(&u.id); err != nil {
			t.Fatal(err)
		}
		users = append(users, u)
	}

	for item := range 25 {
		id := uuid.NewV7().String()
		typ := []string{"game", "movie", "series", "book"}[rng.Intn(4)]
		precision := precisions[rng.Intn(len(precisions))]
		var release *time.Time
		if rng.Intn(5) > 0 {
			d := at.AddDate(0, 0, rng.Intn(400)-300)
			release = &d
		}
		relStatus := []string{"released", "announced", "cancelled"}[rng.Intn(3)]
		if _, err := q.ExecContext(
			ctx,
			`INSERT INTO items (id, type, slug, wikidata_qid, release_date, release_precision, release_status)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			id,
			typ,
			fmt.Sprintf("item-%d", item),
			fmt.Sprintf("Q%d", 1000+item),
			release,
			precision,
			relStatus,
		); err != nil {
			t.Fatal(err)
		}

		var inputs []scoring.Rating
		for _, u := range users {
			if rng.Intn(3) == 0 {
				continue
			}
			var c [6]float64
			for k := range c {
				c[k] = float64(5 * (1 + rng.Intn(20)))
			}
			ratedAt := at.Add(-time.Duration(rng.Intn(350*24)) * time.Hour)
			if _, err := q.ExecContext(
				ctx,
				`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)`,
				u.id,
				id,
				c[0],
				c[1],
				c[2],
				c[3],
				c[4],
				c[5],
				ratedAt,
			); err != nil {
				t.Fatal(err)
			}
			status := statuses[rng.Intn(len(statuses))]
			if status != "" {
				exec(t, q, `INSERT INTO item_status (user_id, item_id, status) VALUES ($1, $2, $3)`, u.id, id, status)
			}
			hasTTB := rng.Intn(2) == 0
			if hasTTB {
				exec(t, q, `INSERT INTO time_to_beat (user_id, item_id, main_minutes) VALUES ($1, $2, 600)`, u.id, id)
			}
			shadow := 1.0
			if rng.Intn(6) == 0 {
				shadow = []float64{0, 0.25, 0.5}[rng.Intn(3)]
				exec(
					t,
					q,
					`INSERT INTO rating_flags (user_id, item_id, shadow_weight, reason) VALUES ($1, $2, $3, 'admin_downweight')`,
					u.id,
					id,
					shadow,
				)
			}
			inputs = append(inputs, scoring.Rating{C: c, Weight: scoring.Weight(scoring.Factors{
				AccountAge:    at.Sub(u.created),
				Status:        status,
				IsGame:        typ == "game",
				HasTimeToBeat: hasTTB,
				Shadow:        shadow,
				Banned:        u.banned,
				BeforeRelease: scoring.RatedBeforeRelease(ratedAt, release, precision, relStatus),
			})})
		}

		if err := scoring.Recompute(ctx, q, id, at, p); err != nil {
			t.Fatal(err)
		}
		want := scoring.Compute(inputs, p)

		var n int
		var w float64
		var score, se sql.NullFloat64
		var conf sql.NullString
		if err := q.QueryRowContext(ctx, `SELECT rating_count, rating_weight, score, score_se, score_confidence FROM items WHERE id = $1`, id).
			Scan(&n, &w, &score, &se, &conf); err != nil {
			t.Fatal(err)
		}
		if n != want.N || math.Abs(w-want.W) > 1e-3 { // rating_weight is numeric(12,3)
			t.Errorf("item %d: N/W sql %d/%v go %d/%v", item, n, w, want.N, want.W)
		}
		if want.N > 0 && math.Abs(score.Float64-want.B) > 1e-4 { // score is numeric(7,4)
			t.Errorf("item %d: B sql %v go %v", item, score.Float64, want.B)
		}
		if conf.String != want.Confidence {
			t.Errorf("item %d: confidence sql %q go %q", item, conf.String, want.Confidence)
		}
	}
}

func exec(t *testing.T, q *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := q.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatal(err)
	}
}
