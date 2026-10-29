package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// GuardName is the guard name of /api/v1 (ratelimit.ByUser, auth.User).
const GuardName = "apikey"

// Guard authenticates /api/v1 requests by an hx_live_ key in `Authorization: Bearer` (SDD 05 §6.1). It accepts
// nothing else: an hx_app_ token or a session is refused (FR-05-13). It is an auth.APIGuard, so auth.Required
// answers problem+json 401 with WWW-Authenticate, and an auth.FailClosedChecker, so a store outage is 503.
type Guard struct {
	Keys *Keys
	// User loads the key's owner by id (accounts.Users.ByID).
	User func(ctx context.Context, id string) (models.User, error)
}

type keyIDKey struct{}

// Name implements auth.Checker.
func (*Guard) Name() string { return GuardName }

// APIOnly marks the guard as an auth.APIGuard.
func (*Guard) APIOnly() {}

// Check implements auth.Checker.
func (g *Guard) Check(c *web.Ctx) bool { return g.Authenticate(c) == nil }

// Authenticate implements auth.FailClosedChecker: nil when authenticated, auth.ErrUnauthenticated for a missing,
// malformed, unknown or revoked key or a deleted owner, and a 503 error when the lookup fails.
func (g *Guard) Authenticate(c *web.Ctx) error {
	if _, err := auth.User[models.User](c, GuardName); err == nil {
		return nil
	}

	raw, ok := auth.BearerToken(c.Request())
	if !ok {
		return auth.ErrUnauthenticated
	}

	keyID, userID, err := g.Keys.Resolve(c.Context(), raw)
	if errors.Is(err, ErrUnknownKey) {
		web.SecurityEvent(c.Request(), "apikey.rejected", "guard", GuardName)
		return auth.ErrUnauthenticated
	}

	if err != nil {
		return Unavailable(fmt.Errorf("resolve api key: %w", err))
	}

	u, err := g.User(c.Context(), strconv.FormatInt(userID, 10))
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) || errors.Is(err, auth.ErrUnauthenticated) ||
			errors.Is(err, orm.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return auth.ErrUnauthenticated
		}

		return Unavailable(fmt.Errorf("load api key owner: %w", err))
	}

	auth.SetUser(c, GuardName, u)
	c.SetRequest(c.Request().WithContext(context.WithValue(c.Context(), keyIDKey{}, keyID)))

	return nil
}

// KeyID returns the id of the key that authenticated c (quota metering, P1.H.1.2).
func KeyID(c *web.Ctx) (string, bool) {
	id, ok := c.Context().Value(keyIDKey{}).(string)
	return id, ok
}

// Unavailable is the fail-closed 503 of a store outage, with Retry-After: 30 (SDD 05 NFR-05-3).
func Unavailable(err error) error {
	return errs.Wrap(http.StatusServiceUnavailable, err).WithHeader("Retry-After", "30")
}
