package seeders

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
)

// ErrCommunityLoadProduction is returned when CommunityLoad runs against a production app.
var ErrCommunityLoadProduction = errors.New("CommunityLoad is a load-test seeder and never runs in production")

// CommunityLoad tops the moderation queue up to Count open (submitted) "create" suggestions, written by the
// DemoCatalog users, for qa/k6/moderation.js (P4.B.5.1: 10,000 open suggestions). It is dev-only: it is not in
// All(), bootstrap registers it outside production, and SeedApp refuses a production app. With the default
// Count of 0 a plain `db:seed` does nothing. Run: db:seed CommunityLoad --count 10000.
type CommunityLoad struct {
	Count int
}

// SeedFlags implements vukem.FlagSeeder.
func (s *CommunityLoad) SeedFlags(fs *flag.FlagSet) {
	fs.IntVar(&s.Count, "count", 0, "CommunityLoad: open suggestions to have in the queue (0 seeds nothing)")
}

// SeedApp implements vukem.AppSeeder.
func (s *CommunityLoad) SeedApp(ctx context.Context, app *vukem.App) error {
	if app.Config.IsProduction() {
		return ErrCommunityLoadProduction
	}

	db, err := app.DB()
	if err != nil {
		return err
	}

	return s.Seed(ctx, db)
}

// Seed inserts Count minus the load suggestions already there, so a second run adds nothing. The rows are
// submitted within the last 30 h, oldest first by id order, and unclaimed.
func (s *CommunityLoad) Seed(ctx context.Context, db *database.DB) error {
	if db.Dialect().Name() != "postgres" {
		return errors.New("CommunityLoad needs PostgreSQL")
	}

	return db.Tx(ctx, func(tx *database.Tx) error {
		var have int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM suggestions WHERE meta @> '{"load":true}'`).
			Scan(&have); err != nil {
			return err
		}

		missing := s.Count - have
		if missing <= 0 {
			return nil
		}

		var authors int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE role = 'user'`).Scan(&authors); err != nil {
			return err
		}

		if authors == 0 {
			return errors.New("CommunityLoad needs the DemoCatalog users: run db:seed DemoCatalog first")
		}

		_, err := tx.ExecContext(ctx, `
WITH authors AS (SELECT array_agg(id ORDER BY id) AS ids FROM users WHERE role = 'user'),
ins AS (
  INSERT INTO suggestions (id, author_id, kind, target_type, target_domain, locale, changes, status, submitted_at, meta)
  SELECT gen_random_uuid(), a.ids[1 + (g % cardinality(a.ids))], 'create', 'item',
    (ARRAY['game','movie','series','book'])[1 + (g % 4)],
    (ARRAY['en','pt-BR','es'])[1 + (g % 3)],
    jsonb_build_object('title', jsonb_build_object('after', 'Load suggestion ' || (g + $2))),
    'submitted', now() - make_interval(secs => g * 10), '{"load":true}'
  FROM authors a, generate_series(1, $1::int) g
  RETURNING id, author_id, submitted_at
)
INSERT INTO suggestion_events (suggestion_id, from_status, to_status, actor_id, actor_role, created_at)
SELECT id, 'draft', 'submitted', author_id, 'author', submitted_at FROM ins`, missing, have)
		if err != nil {
			return fmt.Errorf("CommunityLoad: insert: %w", err)
		}

		return nil
	})
}
