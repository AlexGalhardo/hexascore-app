package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/logging"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// BannedProblem writes the 403 `account-banned` problem (RFC 9457) with the
// reason and expiry, which the banned user is entitled to see (SDD 03 §6.2).
func BannedProblem(c *web.Ctx, problemBase string, u models.User) error {
	body := map[string]any{
		"type":   strings.TrimSuffix(problemBase, "/") + "/account-banned",
		"title":  "Account banned",
		"status": http.StatusForbidden,
		"code":   "account-banned",
	}
	if u.BanReason != nil {
		body["reason"] = *u.BanReason
	}

	if u.BanExpiresAt.Valid {
		body["expiresAt"] = u.BanExpiresAt.V.UTC().Format(time.RFC3339)
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return c.Blob(http.StatusForbidden, "application/problem+json", raw)
}

// NotBanned answers 403 `account-banned` to a signed-in user whose ban is
// active. A ban ends sessions and tokens in its own transaction; this also
// stops survivors and bans written by plain SQL (SDD 03 §8). user returns
// the signed-in user or an error when nobody is (those requests pass).
func NotBanned(
	user func(c *web.Ctx) (models.User, error), problemBase string, now func() time.Time,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := web.NewCtx(w, r, web.ServicesFrom(r.Context()))

			u, err := user(c)
			if err == nil {
				logging.SetUser(r.Context(), u.AuthID())
			}

			if err == nil && u.IsBanned(now()) {
				web.SecurityEvent(r, "auth.banned_request", "user", u.ID)

				if err := BannedProblem(c, problemBase, u); err != nil {
					c.HandleError(err)
				}

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
