package bootstrap

import (
	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// scoringProvider registers the scoring listeners, which need the booted app (database and queue).
type scoringProvider struct{}

// Register subscribes scoring to the account events.
func (scoringProvider) Register(app *vukem.App) error {
	app.Events.Register(scoring.Listeners(app)...)

	return nil
}
