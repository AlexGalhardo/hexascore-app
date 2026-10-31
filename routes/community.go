package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/community"
	ctl "github.com/AlexGalhardo/hexascore-app/app/controllers/community"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

// Community registers the suggest page and the draft save routes (SDD 04 §6). Submit, withdraw, uploads and the
// status pages belong to later rows.
func Community(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	ff, err := app.Flags()
	if err != nil {
		return err
	}

	h := ctl.NewSuggest(acc.DB, ff, acc.Web.Name(), app.Now)
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)
	burst := community.DraftBurst(limiters, acc.Web.Name())

	r.Get("/suggest/{target_type}/{id}", h.Edit).Name("community.suggest").Use(required, verified)
	r.Post("/suggestions", h.Create).Name("community.suggestions.create").Use(required, verified, burst)
	r.Patch("/suggestions/{id}", h.Update).Name("community.suggestions.update").Use(required, verified, burst)

	return nil
}
