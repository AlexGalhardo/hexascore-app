package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// NicknameRequired refuses writes (anything but GET, HEAD and OPTIONS) from
// a signed-in user who has not chosen a nickname: web requests are sent to
// /onboarding, JSON requests get 403 `nickname-required` (SDD 03 FR-03-9).
// Apply it to member writes, never to /onboarding itself. user returns the
// signed-in user or an error when nobody is; hasNickname reads the column.
func NicknameRequired(
	user func(c *web.Ctx) (models.User, error),
	hasNickname func(c *web.Ctx, userID int64) (bool, error),
	problemBase string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			c := web.NewCtx(w, r, web.ServicesFrom(r.Context()))

			u, err := user(c)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			has, err := hasNickname(c, u.ID)
			if err != nil {
				c.HandleError(err)
				return
			}

			if has {
				next.ServeHTTP(w, r)
				return
			}

			if err := nicknameRequired(c, problemBase); err != nil {
				c.HandleError(err)
			}
		})
	}
}

func nicknameRequired(c *web.Ctx, problemBase string) error {
	// Inertia visits ask for JSON but follow redirects.
	if c.Request().Header.Get("X-Inertia") != "" || !c.WantsJSON() {
		return c.Redirect("/onboarding")
	}

	raw, err := json.Marshal(map[string]any{
		"type":   strings.TrimSuffix(problemBase, "/") + "/nickname-required",
		"title":  "Nickname required",
		"status": http.StatusForbidden,
		"code":   "nickname-required",
	})
	if err != nil {
		return err
	}

	return c.Blob(http.StatusForbidden, "application/problem+json", raw)
}
