package console

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"maps"
	"slices"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// GenresCanonicalize is `genres:canonicalize`: links every item to its canonical genres (one translated list per
// category, owner 2026-11-02) and hides the source genre links that fit none. Idempotent; --dry-run prints the
// report and changes nothing. Production: dry run, real run, dry run again, with the owner's yes.
func GenresCanonicalize() vukem.Command {
	var (
		dryRun *bool
		top    *int
	)

	return vukem.CommandFunc{
		CommandName: "genres:canonicalize",
		Desc:        "Replace source genres (Wikidata labels) with the canonical translated genres",
		DefineFlags: func(fs *flag.FlagSet) {
			dryRun = fs.Bool("dry-run", false, "print the report and change nothing")
			top = fs.Int("top", 40, "unmapped labels to list, most used first")
		},
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}

			rep, err := catalog.CanonicalizeGenres(ctx, db, *dryRun)
			out := app.Stdout()
			_, _ = fmt.Fprintf(out,
				"Canonical genres created: %d\nCanonical genres adopted: %d\nSource links: %d\nMapped: %d\nHidden: %d\nDry run: %t\n",
				rep.Created, rep.Adopted, rep.Links, rep.Mapped, rep.Hidden, *dryRun)

			labels := slices.SortedFunc(maps.Keys(rep.Unmapped), func(a, b string) int {
				return cmp.Or(cmp.Compare(rep.Unmapped[b], rep.Unmapped[a]), cmp.Compare(a, b))
			})
			if len(labels) > 0 {
				_, _ = fmt.Fprintf(out, "Unmapped labels (%d):\n", len(labels))
			}
			for _, l := range labels[:min(*top, len(labels))] {
				_, _ = fmt.Fprintf(out, "  %6d  %s\n", rep.Unmapped[l], l)
			}

			return err
		},
	}
}
