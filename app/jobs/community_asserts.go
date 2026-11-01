package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// CommunityAsserts runs the SDD 04 §10 nightly asserts. A failure logs `alert=community_asserts` at ERROR (SDD 12
// §6.2) and fails the job, which alert A6 reports.
type CommunityAsserts struct{}

// Kind names the job on the queue.
func (CommunityAsserts) Kind() string { return "community.asserts" }

func handleCommunityAsserts(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ CommunityAsserts) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		err = communityAsserts(ctx, db)
		if err != nil {
			app.Log.ErrorContext(ctx, "community asserts failed", "alert", "community_asserts", "err", err)
		}

		return err
	})
}

var communityChecks = []struct{ name, sql string }{
	{"applied without audit rows", `SELECT count(*) FROM suggestions s WHERE s.status = 'applied'
  AND NOT EXISTS (SELECT 1 FROM catalog_audit a WHERE a.suggestion_id = s.id)`},
	{"open on a missing target", `SELECT count(*) FROM suggestions s
WHERE s.status IN ('draft','submitted','in_review','approved') AND s.target_id IS NOT NULL AND NOT CASE s.target_type
  WHEN 'item' THEN EXISTS (SELECT 1 FROM items WHERE id = s.target_id)
  WHEN 'edition' THEN EXISTS (SELECT 1 FROM editions WHERE id = s.target_id)
  ELSE EXISTS (SELECT 1 FROM entities WHERE id = s.target_id) END`},
}

func communityAsserts(ctx context.Context, db *database.DB) error {
	var failed []string

	for _, c := range communityChecks {
		var n int64
		if err := db.QueryRowContext(ctx, c.sql).Scan(&n); err != nil {
			return fmt.Errorf("community asserts: %s: %w", c.name, err)
		}

		if n > 0 {
			failed = append(failed, fmt.Sprintf("%d %s", n, c.name))
		}
	}

	n, err := community.StatsDrift(ctx, db)
	if err != nil {
		return err
	}

	if n > 0 {
		failed = append(failed, fmt.Sprintf("%d stats differ from rebuild", n))
	}

	if len(failed) > 0 {
		return errors.New("community asserts: " + strings.Join(failed, "; "))
	}

	return nil
}
