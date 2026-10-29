package console

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// ErrCatalogCheckFailed is returned when an error-level data-quality assert
// has violations, so the process exits non-zero (SDD 01 §7.6).
var ErrCatalogCheckFailed = errors.New("catalog:check found error-level violations")

// CatalogCheck is `catalog:check [--json]`: runs the DQ-01..DQ-14 asserts,
// prints counts (or the JSON report) and fails on any error-level violation.
func CatalogCheck() vukem.Command {
	var asJSON *bool

	return vukem.CommandFunc{
		CommandName: "catalog:check",
		Desc:        "Run the catalogue data-quality asserts; exit 1 on any error",
		DefineFlags: func(fs *flag.FlagSet) { asJSON = fs.Bool("json", false, "print the report as JSON") },
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}
			report, err := catalog.RunDQ(ctx, db)
			if err != nil {
				return fmt.Errorf("catalog:check: %w", err)
			}

			out := app.Stdout()
			if *asJSON {
				if err := json.NewEncoder(out).Encode(report); err != nil {
					return err
				}
			} else {
				for _, c := range report.Checks {
					_, _ = fmt.Fprintf(out, "%s %-5s %6d  %s\n", c.ID, c.Level, c.Count, c.Description)
				}
				_, _ = fmt.Fprintf(out, "errors: %d, warnings: %d\n", report.Errors, report.Warnings)
			}
			if report.Errors > 0 {
				return ErrCatalogCheckFailed
			}
			return nil
		},
	}
}
