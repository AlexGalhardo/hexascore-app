package bootstrap

import (
	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// socialProvider records feed activity from the rating, status and list events (SDD 14 §12, ADR 0021).
type socialProvider struct{}

// Register subscribes the activity recorder; the tables exist only on Postgres (VF-033).
func (socialProvider) Register(app *vukem.App) error {
	db, err := app.DB()
	if err != nil {
		return err
	}

	if db.Dialect().Name() != "postgres" {
		return nil
	}

	ff, err := app.Flags()
	if err != nil {
		return err
	}

	app.Events.Register((&social.Follows{DB: db, Flags: ff}).Listeners()...)

	return nil
}
