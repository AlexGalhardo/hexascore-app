// Package developer serves the API key management page /profile/api (SDD 09 §6.2, SDD 05 FR-05-2).
package developer

import (
	"errors"
	"net/http"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// ErrorLimit is the validation code on `name` when the user already has api.MaxLiveKeys live keys; the page
// translates it (api.profile.keys.errors.limit).
const ErrorLimit = "limit"

// APIKeys replaces the legacy createApiKeyAction and revokeApiKeyAction.
type APIKeys struct {
	Keys  *api.Keys
	Guard string
}

func (h *APIKeys) user(c *web.Ctx) (models.User, error) { return auth.User[models.User](c, h.Guard) }

// Index renders Developer/ApiKeys with every key of the user. The plain key of a key created on the previous
// request arrives only as the `newKey` flash, so it is shown once.
func (h *APIKeys) Index(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	keys, err := h.Keys.List(c.Context(), u.ID)
	if err != nil {
		return err
	}

	return inertia.Render(c, "Developer/ApiKeys", inertia.Props{"keys": keys, "maxKeys": api.MaxLiveKeys})
}

// Create stores a key and flashes its plain value.
func (h *APIKeys) Create(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.APIKeySchema)
	if err != nil {
		return err
	}

	plain, _, err := h.Keys.Create(c.Context(), u.ID, in.Name)
	if errors.Is(err, api.ErrTooManyKeys) {
		return errs.NewValidation().Add("name", ErrorLimit)
	}

	if err != nil {
		return err
	}

	inertia.Flash(c, "newKey", plain)

	return c.Redirect("/profile/api")
}

// Revoke revokes one of the user's keys; another user's key answers 404 like a missing one.
func (h *APIKeys) Revoke(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	err = h.Keys.Revoke(c.Context(), u.ID, c.Param("id"))
	if errors.Is(err, api.ErrKeyNotFound) {
		return errs.NewKey(http.StatusNotFound, "api.keys.not_found", "Not found.", nil).
			WithExtension("code", "not-found")
	}

	if err != nil {
		return err
	}

	return c.Redirect("/profile/api")
}
