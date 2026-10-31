package account

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

func readCeremony(c *web.Ctx) ([]byte, error) {
	return io.ReadAll(io.LimitReader(c.Request().Body, 1<<16))
}

// PasskeyLoginOptions starts a discoverable sign-in.
func (h *Handlers) PasskeyLoginOptions(c *web.Ctx) error {
	opts, err := h.a.BeginPasskeyLogin(c.Context(), c.Session())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, opts)
}

// PasskeyLogin verifies an assertion, skips TOTP, and signs the user in.
func (h *Handlers) PasskeyLogin(c *web.Ctx) error {
	body, err := readCeremony(c)
	if err != nil {
		return err
	}

	u, err := h.a.FinishPasskeyLogin(c.Context(), c.Session(), body)
	if err != nil {
		if banned, ok := errors.AsType[*accounts.PasskeyBannedError](err); ok {
			return middleware.BannedProblem(c, h.a.ProblemBase, banned.User)
		}

		return err
	}

	if err := h.a.Web.Login(c, u, auth.Remember(false)); err != nil {
		return err
	}

	return h.a.RedirectAfterLogin(c, u)
}

// PasskeyOptions starts registration after a password or fresh Google confirmation.
func (h *Handlers) PasskeyOptions(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.PasskeyOptionsSchema)
	if err != nil {
		return err
	}

	opts, err := h.a.BeginPasskeyRegistration(c.Context(), c.Session(), u, in.Name, in.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, opts)
}

// PasskeyCreate stores a verified registration and returns to the profile.
func (h *Handlers) PasskeyCreate(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	body, err := readCeremony(c)
	if err != nil {
		return err
	}

	if err := h.a.FinishPasskeyRegistration(c.Context(), c.Session(), u, body); err != nil {
		return err
	}

	return c.Redirect("/profile")
}

// PasskeyDelete removes one passkey owned by the signed-in user.
func (h *Handlers) PasskeyDelete(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return h.a.PasskeyMissing(c.Context())
	}

	if err := h.a.RemovePasskey(c.Context(), u.ID, id); err != nil {
		return err
	}

	return c.Redirect("/profile")
}
