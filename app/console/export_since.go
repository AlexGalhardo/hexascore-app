package console

import (
	"context"
	"flag"
	"fmt"
	"time"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/database/legacyexport"
)

// ExportSince is `hexascore:export-since <RFC3339>` (SDD 11 §7.6 R3). It prints
// one legacy upsert per line. scripts/legacy/replay.sh applies that SQL after
// Neon is writable again. It does not open a connection to Neon.
func ExportSince() vukem.Command {
	var sinceFlag *string

	return vukem.CommandFunc{
		CommandName: "hexascore:export-since",
		Desc:        "Print legacy upserts for rows written since the cutover (RFC3339). Pipe to scripts/legacy/replay.sh",
		DefineFlags: func(fs *flag.FlagSet) {
			sinceFlag = fs.String("since", "", "cutover instant, RFC3339 (or pass it as the first argument)")
		},
		Fn: func(ctx context.Context, app *vukem.App, args []string) error {
			raw := *sinceFlag
			if raw == "" && len(args) > 0 {
				raw = args[0]
			}

			since, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return fmt.Errorf("hexascore:export-since: cutover must be RFC3339, got %q", raw)
			}

			db, err := app.DB()
			if err != nil {
				return err
			}

			lines, err := legacyexport.Statements(ctx, db, since)
			if err != nil {
				return err
			}

			out := app.Stdout()
			if len(lines) == 0 {
				_, _ = fmt.Fprintln(out, "-- export-since: no rows")
				return nil
			}

			for _, line := range lines {
				if _, err := fmt.Fprintln(out, line); err != nil {
					return fmt.Errorf("hexascore:export-since: write: %w", err)
				}
			}

			return nil
		},
	}
}
