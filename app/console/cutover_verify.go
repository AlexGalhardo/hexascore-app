package console

import (
	"context"
	"flag"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/drivers/s3"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

// CutoverVerify is `hexascore:cutover-verify [--counts-only] [--sample-pct N]`
// (SDD 11 §7.6 step 6): per-table counts, id-sample hashes, cover keys in the
// bucket (S3_*), orphan foreign keys and the score recompute, between schema
// legacy_import and the imported tables. It only reads; any failed check
// exits 1 naming the tables.
func CutoverVerify() vukem.Command {
	var (
		countsOnly *bool
		samplePct  *int
	)

	return vukem.CommandFunc{
		CommandName: "hexascore:cutover-verify",
		Desc:        "Verify the legacy import before the DNS switch (counts, samples, covers, orphans, scores)",
		DefineFlags: func(fs *flag.FlagSet) {
			countsOnly = fs.Bool("counts-only", false, "run only the per-table counts")
			samplePct = fs.Int("sample-pct", 1, "percent of ids hashed per table (1-100)")
		},
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			if *samplePct < 1 || *samplePct > 100 {
				return fmt.Errorf("hexascore:cutover-verify: --sample-pct must be 1-100, got %d", *samplePct)
			}

			db, err := app.DB()
			if err != nil {
				return err
			}

			o := legacyimport.VerifyOptions{CountsOnly: *countsOnly, SamplePct: *samplePct, Out: app.Stdout()}

			if !*countsOnly {
				cfg, err := config.Load[s3.Config](app.Source)
				if err != nil {
					return fmt.Errorf("hexascore:cutover-verify: S3 config: %w", err)
				}

				if o.Disk, err = s3.New(cfg); err != nil {
					return fmt.Errorf("hexascore:cutover-verify: S3 disk: %w", err)
				}
			}

			return legacyimport.Verify(ctx, db, o)
		},
	}
}
