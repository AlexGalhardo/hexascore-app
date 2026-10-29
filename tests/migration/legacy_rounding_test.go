package migration_test

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

// TestLegacyRounding pins FR-02-13: legacy 0-100 categories become the nearest multiple of 5 in 5..100,
// and the social stage applies the same rule in SQL with its audit trail.
func TestLegacyRounding(t *testing.T) {
	t.Parallel()

	want := map[int]int{0: 5, 1: 5, 2: 5, 3: 5, 12: 10, 13: 15, 97: 95, 98: 100, 100: 100}
	for in, out := range want {
		if got := legacyimport.RoundRating(in); got != out {
			t.Errorf("RoundRating(%d) = %d, want %d", in, got, out)
		}
	}

	t.Run("import", func(t *testing.T) {
		t.Parallel()

		if os.Getenv("TEST_DATABASE_URL") == "" {
			t.Skip("TEST_DATABASE_URL unset: the import half needs PostgreSQL (docker compose up -d --wait)")
		}

		app := vukemtest.New(t, bootstrap.New,
			vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})

		db, err := app.DB()
		if err != nil {
			t.Fatal(err)
		}

		fixture, err := os.ReadFile("../../database/legacyimport/testdata/fixture.sql")
		if err != nil {
			t.Fatal(err)
		}

		ctx := t.Context()
		if _, err := db.ExecContext(ctx, string(fixture)); err != nil {
			t.Fatalf("load fixture: %v", err)
		}

		// A spread of values over 0..100 from u06, so the SQL rule is checked against RoundRating.
		if _, err := db.ExecContext(ctx, `INSERT INTO legacy_import.ratings (user_id, item_id, c1, c2, c3, c4, c5, c6)
SELECT 'u06', i.id, 21 * g, 21 * g + 3, 21 * g + 6, 21 * g + 9, 21 * g + 14, 100 - 21 * g FROM generate_series(0, 4) g
JOIN (SELECT id, row_number() OVER (ORDER BY id) - 1 AS n FROM legacy_import.items) i ON i.n = g`); err != nil {
			t.Fatal(err)
		}

		stages, err := legacyimport.Select(legacyimport.Stages(), []string{"users", "catalog", "social"})
		if err != nil {
			t.Fatal(err)
		}

		r := &legacyimport.Runner{DB: db, Batch: 3, Out: io.Discard, Now: time.Now}
		if err := r.Run(ctx, stages); err != nil {
			t.Fatalf("run: %v", err)
		}

		rows, err := db.QueryContext(
			ctx,
			`SELECT l.c1, l.c2, l.c3, l.c4, l.c5, l.c6, n.c1, n.c2, n.c3, n.c4, n.c5, n.c6,
  a.user_id IS NOT NULL
FROM legacy_import.ratings l JOIN users u ON u.legacy_id = l.user_id
JOIN ratings n ON n.user_id = u.id AND n.item_id = l.item_id
LEFT JOIN rating_migration_audit a ON a.user_id = n.user_id AND a.item_id = n.item_id`,
		)
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = rows.Close() }()

		changed, audited, read := 0, 0, 0

		for rows.Next() {
			var (
				oldC, newC [6]int
				hasAudit   bool
			)

			if err := rows.Scan(&oldC[0], &oldC[1], &oldC[2], &oldC[3], &oldC[4], &oldC[5],
				&newC[0], &newC[1], &newC[2], &newC[3], &newC[4], &newC[5], &hasAudit); err != nil {
				t.Fatal(err)
			}

			read++

			for k := range oldC {
				if want := legacyimport.RoundRating(oldC[k]); newC[k] != want {
					t.Errorf("legacy %v imported as %v: c%d want %d", oldC, newC, k+1, want)
				}
			}

			if oldC != newC {
				changed++
			}

			if hasAudit {
				audited++
			}
		}

		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}

		if read != 8 || audited != changed {
			t.Fatalf(
				"read %d ratings, %d changed, %d audit rows; want 8 read and audit rows = changed",
				read,
				changed,
				audited,
			)
		}

		var summaries int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM score_actions WHERE kind = 'legacy_rounding'
  AND (after->>'rows_changed')::int = $1`, changed).Scan(&summaries); err != nil || summaries != 1 {
			t.Fatalf("legacy_rounding summaries with rows_changed=%d: %d (%v), want 1", changed, summaries, err)
		}
	})
}
