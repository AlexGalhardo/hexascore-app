package console

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strconv"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// IngestBootstrap is `ingest:bootstrap [--limit N]`: queues the bootstrap chain (SDD 06 §7.6), whose stages
// follow one another on the ingest queue. --limit caps the Wikidata top-N per type (0 keeps the default).
func IngestBootstrap() vukem.Command {
	var limit *int

	return vukem.CommandFunc{
		CommandName: "ingest:bootstrap",
		Desc:        "Queue the bootstrap ingestion chain: Wikidata top-N, entities, normalise",
		DefineFlags: func(fs *flag.FlagSet) { limit = fs.Int("limit", 0, "Wikidata top-N per type; 0 keeps the default") },
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			if *limit < 0 {
				return errors.New("ingest:bootstrap: --limit must not be negative")
			}

			db, err := app.DB()
			if err != nil {
				return err
			}

			q, err := app.Queue()
			if err != nil {
				return err
			}

			var params map[string]json.RawMessage
			if *limit > 0 {
				params = map[string]json.RawMessage{"limit": json.RawMessage(strconv.Itoa(*limit))}
			}

			r, err := ingest.Bootstrap.Start(ctx, db, q, "cli", params)
			if err != nil {
				return fmt.Errorf("ingest:bootstrap: %w", err)
			}

			_, _ = fmt.Fprintf(
				app.Stdout(), "Queued run %d (%s); run `vukem queue:work -queue ingest` to process it\n", r.ID, r.Job,
			)

			return nil
		},
	}
}
