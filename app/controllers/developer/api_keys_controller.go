// Package developer serves the API key actions of the /profile page (SDD 09 §6.2, SDD 05 FR-05-2).
package developer

import (
	"errors"
	"net/http"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// Path is where the key management lives: a section of /profile (the page /profile/api answers 301 here).
const Path = "/profile#api"

// ErrorLimit is the validation code on `name` when the user already has api.MaxLiveKeys live keys; the page
// translates it (api.profile.keys.errors.limit).
const ErrorLimit = "limit"

// APIKeys replaces the legacy createApiKeyAction and revokeApiKeyAction.
type APIKeys struct {
	Keys  *api.Keys
	Guard string
}

func (h *APIKeys) user(c *web.Ctx) (models.User, error) { return auth.User[models.User](c, h.Guard) }

// Index redirects the old key page to its section of /profile; bookmarks and docs links keep working.
func (*APIKeys) Index(c *web.Ctx) error {
	return c.RedirectStatus(http.StatusMovedPermanently, i18n.PathPrefix(c.Context())+Path)
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

	return c.Redirect("/profile")
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

	return c.Redirect("/profile")
}
