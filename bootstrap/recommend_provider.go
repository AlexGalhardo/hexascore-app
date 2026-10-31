package bootstrap

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/events"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/primitives"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// recommendProvider feeds search clicks to the trending counter (SDD 08 open issue 08-5), queues rebuilds on bans.
type recommendProvider struct{}

// Register subscribes to search.ResultClicked; only the item id is used.
func (recommendProvider) Register(app *vukem.App) error {
	rec := primitives.For(app)
	app.Events.Register(events.On(func(_ context.Context, e search.ResultClicked) error {
		rec.CountClick(e.ItemID, e.At)

		return nil
	}))

	app.Events.Register(trending.Listeners(app)...)

	return nil
}
