package console

import (
	"context"
	"flag"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
)

// CoversRekey is `covers:rekey`: copies every item cover from its content-hashed key to its slug key
// (covers/{section}/{slug}[-t].webp, owner decision 2026-11-01, item 18) on the S3_* bucket and points the row at
// it. Idempotent; --dry-run counts and changes nothing. The old objects are deleted by covers.gc 30 days later.
func CoversRekey() vukem.Command {
	var (
		dryRun *bool
		batch  *int
	)

	return vukem.CommandFunc{
		CommandName: "covers:rekey",
		Desc:        "Move item covers to their slug keys (covers/{section}/{slug}.webp) on the bucket",
		DefineFlags: func(fs *flag.FlagSet) {
			dryRun = fs.Bool("dry-run", false, "print the counts, copy and change nothing")
			batch = fs.Int("batch", covers.RekeyBatch, "items per page")
		},
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}

			disk, err := jobs.CoversDisk(app)
			if err != nil {
				return err
			}

			res, err := covers.Rekey(ctx, db, disk, *dryRun, *batch)
			_, _ = fmt.Fprintf(
				app.Stdout(),
				"Items with covers: %d\nAlready on slug keys: %d\nMoved: %d\nObjects copied: %d\nMissing objects: %d\nDry run: %t\n",
				res.Items,
				res.Current,
				res.Moved,
				res.Copied,
				res.Missing,
				*dryRun,
			)

			return err
		},
	}
}
