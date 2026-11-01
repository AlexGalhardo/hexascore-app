package console

import (
	"context"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// SearchReindex is `search:reindex`: rewrites every search index and rebuilds the suggest rows (SDD 07; runbook R6).
func SearchReindex() vukem.Command {
	return vukem.CommandFunc{
		CommandName: "search:reindex",
		Desc:        "Rewrite the search indexes and rebuild the suggest rows",
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}

			if err := search.Reindex(ctx, db); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(app.Stdout(), "Search indexes rewritten and suggest rows rebuilt")

			return nil
		},
	}
}
