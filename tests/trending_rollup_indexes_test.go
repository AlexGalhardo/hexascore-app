//go:build integration

package tests

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestTrendingRollupIndexes loads a year of events with the FK checks off (replica role) and asserts the planner picks
// the new index for each one-day window the hourly rollup reads (FIX-trending-rollup-indexes).
func TestTrendingRollupIndexes(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	tx, err := db.SQL().BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	const day = `now() - (g % 365) * interval '1 day'`
	setup := []string{
		`SET LOCAL session_replication_role = replica`,
		`INSERT INTO rating_events (at, kind, user_id, item_id, item_type)
		 SELECT ` + day + `, 'rating_created', g, gen_random_uuid(), 'game' FROM generate_series(1, 100000) g`,
		`INSERT INTO item_status (user_id, item_id, status, updated_at)
		 SELECT g, gen_random_uuid(), 'want', ` + day + ` FROM generate_series(1, 100000) g`,
		`INSERT INTO list_items (list_id, item_id, position, added_at)
		 SELECT gen_random_uuid(), gen_random_uuid(), 1, ` + day + ` FROM generate_series(1, 100000) g`,
		`INSERT INTO suggestion_events (suggestion_id, to_status, actor_role, created_at)
		 SELECT gen_random_uuid(), 'submitted', 'author', ` + day + ` FROM generate_series(1, 100000) g`,
		`ANALYZE rating_events, item_status, list_items, suggestion_events`,
	}
	for _, q := range setup {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	for _, c := range []struct{ table, column, index string }{
		{"rating_events", "at", "rating_events_at_idx"},
		{"item_status", "updated_at", "item_status_updated_at_idx"},
		{"list_items", "added_at", "list_items_added_at_idx"},
		{"suggestion_events", "created_at", "suggestion_events_created_at_idx"},
	} {
		out := explain(t, tx, `EXPLAIN SELECT 1 FROM `+c.table+` WHERE `+c.column+
			` >= date_trunc('day', now()) - interval '10 days' AND `+c.column+` < date_trunc('day', now()) - interval '9 days'`)
		t.Logf("%s:\n%s", c.table, out)

		if !strings.Contains(out, c.index) {
			t.Errorf("%s: plan does not use %s", c.table, c.index)
		}
	}
}

func explain(t *testing.T, tx *sql.Tx, query string) string {
	t.Helper()

	rows, err := tx.QueryContext(t.Context(), query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var plan []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}

		plan = append(plan, line)
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	return strings.Join(plan, "\n")
}
