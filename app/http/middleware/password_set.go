package middleware

import (
	"context"
	"net/http"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// PasswordSet answers 403 to a signed-in user who never chose a password
// (accounts created through Google hold a random one), so two-factor setup
// cannot be the only credential check of such an account (SDD 03 A07).
// message returns the explanation for the request's locale.
func PasswordSet(
	user func(c *web.Ctx) (models.User, error),
	isSet func(ctx context.Context, userID int64) (bool, error),
	message func(ctx context.Context) string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := web.NewCtx(w, r, web.ServicesFrom(r.Context()))

			u, err := user(c)
			if err != nil {
				c.HandleError(err)

				return
			}

			set, err := isSet(r.Context(), u.ID)
			if err != nil {
				c.HandleError(err)

				return
			}

			if !set {
				c.HandleError(errs.New(http.StatusForbidden, message(r.Context())))

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
