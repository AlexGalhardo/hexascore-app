package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/alerts"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// IngestDigest is the scheduled task `ingest:digest` (every 15 minutes while the catalogue is filled, RFC-001 section 8). It does nothing while the
// ingestion master switch is off, so it ships dark; notifier is read at run time because the Notifier is built when
// the routes are, after the schedule is declared.
func IngestDigest(app *vukem.App, notifier func() *alerts.Notifier) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		ff, err := app.Flags()
		if err != nil {
			return err
		}

		if !run.Allowed(ctx, ff, "") {
			return nil
		}

		db, err := app.DB()
		if err != nil {
			return err
		}

		_, err = alerts.SendIngestDigest(ctx, db, notifier(), app.Now(), false)

		return err
	}
}
