package console

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

// ImportLegacy is `hexascore:import-legacy --stage all|<name>[,<name>] [--dry-run] [--stdin]`
// (SDD 11 §7.6 step 5). It reads the restored dump in schema legacy_import:
// the dry run prints the pre-flight counts and the planned rows per stage and
// writes nothing; a real run aborts on any pre-flight failure, then imports
// the stages in order, resuming each from its checkpoint.
func ImportLegacy() vukem.Command {
	var (
		stage        *string
		dryRun       *bool
		batch        *int
		stdin        *bool
		allowNoLogin *bool
	)

	return vukem.CommandFunc{
		CommandName: "hexascore:import-legacy",
		Desc:        "Import the legacy dump from schema legacy_import (stages, checkpoints, --dry-run)",
		DefineFlags: func(fs *flag.FlagSet) {
			stage = fs.String(
				"stage",
				"all",
				"all, or a comma list of: users,twofactor,catalog,entities,social,library,billing,finalise",
			)
			dryRun = fs.Bool("dry-run", false, "print pre-flight and planned counts, write nothing")
			batch = fs.Int("batch", legacyimport.DefaultBatch, "rows per transaction")
			stdin = fs.Bool("stdin", false, "read the twofactor export (NDJSON) from stdin")
			allowNoLogin = fs.Bool("allow-no-login", false,
				"opt in to importing users who have no password, Google link, or passkey")
		},
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			stages, err := legacyimport.Select(legacyimport.Stages(), strings.Split(*stage, ","))
			if err != nil {
				return err
			}

			if *batch < 1 {
				return errors.New("hexascore:import-legacy: --batch must be at least 1")
			}

			db, err := app.DB()
			if err != nil {
				return err
			}

			out := app.Stdout()

			pre, preErr := legacyimport.Preflight(ctx, db)
			if preErr != nil && !isPreflight(preErr) {
				return preErr
			}

			_, _ = fmt.Fprintf(
				out,
				"pre-flight: duplicate_emails=%d twofactor_without_row=%d no_login_method=%d passkey_only=%d\n",
				pre.Failures["duplicate_emails"],
				pre.Failures["twofactor_without_row"],
				pre.Failures["no_login_method"],
				pre.PasskeyOnly,
			)

			if *allowNoLogin {
				pre, preErr = legacyimport.IgnoreNoLogin(pre, preErr)
			}

			if *dryRun {
				plan, err := legacyimport.Plan(ctx, db, stages)
				if err != nil {
					return err
				}

				for _, p := range plan {
					_, _ = fmt.Fprintf(out, "plan %-9s %-18s %d\n", p.Stage, p.Table, p.Rows)
				}

				_, _ = fmt.Fprintln(out, "dry run: nothing written")

				return preErr
			}

			if preErr != nil {
				return preErr
			}

			// The 2FA secrets arrive only on stdin, decrypted by scripts/legacy/twofactor-export.ts (SDD 03 §7.6).
			for i := range stages {
				if stages[i].Name == "twofactor" {
					if !*stdin {
						return legacyimport.ErrNeedsStdin
					}

					stages[i].Batch = legacyimport.TwoFactorBatch(app.Stdin(), app.Now)
				}
			}

			r := &legacyimport.Runner{DB: db, Batch: *batch, Out: out, Now: app.Now}

			return r.Run(ctx, stages)
		},
	}
}

func isPreflight(err error) bool {
	_, ok := errors.AsType[*legacyimport.PreflightError](err)
	return ok
}
