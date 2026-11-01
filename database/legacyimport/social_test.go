//go:build integration

package legacyimport_test

import (
	"io"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

type check struct{ name, query, want string }

func runStages(t *testing.T, db *database.DB, names ...string) {
	t.Helper()

	stages, err := legacyimport.Select(legacyimport.Stages(), names)
	if err != nil {
		t.Fatal(err)
	}

	r := &legacyimport.Runner{DB: db, Batch: 2, Out: io.Discard, Now: time.Now}
	if err := r.Run(t.Context(), stages); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func assertChecks(t *testing.T, db *database.DB, checks []check) {
	t.Helper()

	for _, c := range checks {
		var got string
		if err := db.QueryRowContext(t.Context(), c.query).Scan(&got); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}

		if got != c.want {
			t.Errorf("%s = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestSocialStage(t *testing.T) {
	t.Parallel()

	_, db := fixtureDB(t)
	runStages(t, db, "users", "catalog", "social")

	assertChecks(t, db, []check{
		{
			"ratings rounded",
			`SELECT string_agg(u.legacy_id || ':' || array_to_string(ARRAY[c1, c2, c3, c4, c5, c6], '/'), ','
  ORDER BY u.legacy_id) FROM ratings r JOIN users u ON u.id = r.user_id`,
			"u01:5/5/5/5/10/15,u02:95/100/100/50/55/60,u05:50/50/50/50/50/50",
		},
		{"one audit row per changed rating", `SELECT count(*)::text FROM rating_migration_audit`, "2"},
		{
			"one legacy_rounding summary",
			`SELECT count(*) || ' ' || min(after->>'rows_read') || ' ' || min(after->>'rows_changed') || ' ' ||
  min(after->'delta_histogram'->'c6'->>'2') FROM score_actions WHERE kind = 'legacy_rounding'`,
			"1 3 2 1",
		},
		{
			"list items dense, added with the list, mixed types kept",
			`SELECT string_agg(i.slug || ':' || li.position || ':' || to_char(li.added_at AT TIME ZONE 'UTC', 'YYYY-MM-DD'), ','
  ORDER BY li.position) FROM list_items li JOIN items i ON i.id = li.item_id
WHERE li.list_id = '01910000-0000-7000-8000-000000000001'`,
			"bare-game:1:2025-01-05,half-life:2:2025-01-05,future-film:3:2025-01-05",
		},
		{
			"lists",
			`SELECT string_agg(title || ':' || type || ':' || is_public || ':' || item_count, ',' ORDER BY title) FROM lists`,
			"Favourites:game:true:3,Private:book:false:0",
		},
		{
			"reports",
			`SELECT string_agg(r.target_type || ':' || coalesce(u.legacy_id, r.target_id) || ':' || r.status, ','
  ORDER BY r.id) FROM reports r LEFT JOIN users u ON r.target_type = 'user' AND u.id::text = r.target_id`,
			"list:01910000-0000-7000-8000-000000000001:open,user:u04:resolved",
		},
		{
			"contact messages", `SELECT string_agg(coalesce(u.legacy_id, '-') || ':' || c.subject, ',' ORDER BY c.id)
FROM contact_messages c LEFT JOIN users u ON u.id = c.user_id`,
			"u01:bug,-:question",
		},
		// Follows stay in the legacy dump (SDD 14 §5).
		{"no follows imported", `SELECT count(*)::text FROM follows`, "0"},
	})

	// A second run of the finished stage writes nothing new.
	runStages(t, db, "social")
	assertChecks(
		t,
		db,
		[]check{{"summary once", `SELECT count(*)::text FROM score_actions WHERE kind = 'legacy_rounding'`, "1"}},
	)
}

func TestLibraryStage(t *testing.T) {
	t.Parallel()

	_, db := fixtureDB(t)
	runStages(t, db, "users", "catalog", "library")

	assertChecks(t, db, []check{
		{
			"item status", `SELECT string_agg(u.legacy_id || ':' || s.status, ',' ORDER BY u.legacy_id)
FROM item_status s JOIN users u ON u.id = s.user_id`,
			"u01:completed,u02:want",
		},
		{
			"time to beat",
			`SELECT u.legacy_id || ':' || t.main_minutes FROM time_to_beat t JOIN users u ON u.id = t.user_id`,
			"u01:600",
		},
	})
}
