package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/developer"
)

// APIKeys registers the key management page /profile/api (legacy path, SDD 09 §6.2). Creating a key needs a
// verified email, as in the legacy action. The /api/v1 group with api.Guard arrives with its first endpoint
// (P1.H.1.2, P1.H.2.1). The table exists only on PostgreSQL (VF-033).
func APIKeys(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	if acc.DB.Dialect().Name() != "postgres" {
		return nil
	}

	h := &developer.APIKeys{Keys: &api.Keys{DB: acc.DB, Now: app.Now}, Guard: acc.Web.Name()}
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)

	r.Get("/profile/api", h.Index).Name("profile.api").Use(required)
	r.Post("/profile/api/keys", h.Create).Name("profile.api.keys.create").Use(required, verified)
	r.Delete("/profile/api/keys/{id}", h.Revoke).Name("profile.api.keys.revoke").Use(required)

	return nil
}
