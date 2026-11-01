package console

import (
	"context"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// IngestBackfillEditions is `ingest:backfill-editions`: queues the one-off edition backfill (SDD 06 §7.8).
func IngestBackfillEditions() vukem.Command {
	return vukem.CommandFunc{
		CommandName: "ingest:backfill-editions",
		Desc:        "Queue the edition backfill: ISBNs and edition QIDs of the stored Wikidata book records",
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}

			q, err := app.Queue()
			if err != nil {
				return err
			}

			r, err := ingest.EditionsBackfill.Start(ctx, db, q, "cli", nil)
			if err != nil {
				return fmt.Errorf("ingest:backfill-editions: %w", err)
			}

			_, _ = fmt.Fprintf(
				app.Stdout(), "Queued run %d (%s); run `vukem queue:work -queue ingest` to process it\n", r.ID, r.Job,
			)

			return nil
		},
	}
}
