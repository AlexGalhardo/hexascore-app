package console

import (
	"context"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// SearchPurge is `search:purge`: deletes search_events older than 30 days
// (SDD 07 §7.8); scheduled daily in routes/schedule.go.
func SearchPurge() vukem.Command {
	return vukem.CommandFunc{
		CommandName: "search:purge",
		Desc:        "Delete search events older than 30 days",
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}

			n, err := search.PurgeEvents(ctx, db, app.Now())
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(app.Stdout(), "Purged search events:", n)

			return nil
		},
	}
}
