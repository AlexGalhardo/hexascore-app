package bootstrap

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/events"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// communityProvider withdraws a banned author's open suggestions (SDD 04 section 8); it needs the booted database.
type communityProvider struct{}

// Register subscribes community to the ban event.
func (communityProvider) Register(app *vukem.App) error {
	app.Events.Register(events.On(func(ctx context.Context, e accounts.UserBanned) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		return community.WithdrawAuthor(ctx, db, e.UserID, "author_banned")
	}))

	return nil
}
