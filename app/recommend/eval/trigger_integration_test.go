//go:build integration

package eval_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/eval"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// OQ-07 (owner, 2026-11-02): flagged seed accounts do not count towards the collaborative-filtering gate.
func TestCheckTriggerIgnoresSeedUsers(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}

	items := make([]string, 10)
	for i := range items {
		items[i] = uuid.NewV7().String()
		exec(`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', $2, $3)`,
			items[i], fmt.Sprintf("g-%d", i), fmt.Sprintf("Q%d", 100+i))
	}

	for i, seed := range []bool{false, true, true} {
		var id int64

		if err := db.QueryRowContext(t.Context(), `INSERT INTO users (name, email, password, is_seed)
VALUES ($1, $2, 'x', $3) RETURNING id`, fmt.Sprintf("u%d", i), fmt.Sprintf("u%d@x.test", i), seed).Scan(&id); err != nil {
			t.Fatal(err)
		}

		for _, item := range items {
			exec(
				`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 50, 50, 50, 50, 50, 50)`,
				id,
				item,
			)
		}
	}

	got, err := eval.CheckTrigger(t.Context(), db, time.Now(), eval.Thresholds{MinUsers: 1, MinRatings: 10})
	if err != nil {
		t.Fatal(err)
	}

	if got.Users != 1 {
		t.Fatalf("users counted = %d, want 1 (the two seed accounts must not count)", got.Users)
	}
}
