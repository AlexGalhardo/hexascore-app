package console

import (
	"context"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// CommunityStatsRebuild is `community:stats:rebuild`. It rewrites contributor_stats from suggestions.
// A correct table is left unchanged.
func CommunityStatsRebuild() vukem.Command {
	return vukem.CommandFunc{
		CommandName: "community:stats:rebuild",
		Desc:        "Rebuild contributor_stats from suggestions and events",
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}
			if err := community.RebuildStats(ctx, db, app.Now()); err != nil {
				return err
			}
			_, err = fmt.Fprintln(app.Stdout(), "contributor_stats rebuilt")

			return err
		},
	}
}
