package middleware

import (
	"context"
	"net/http"

	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// UnreadNotifications shares the unreadNotifications prop (SDD 14 §6.2). It
// is lazy, so requests that render no page never count, and anonymous
// visitors get 0 without a query. A failed count shows 0 instead of failing
// the page: the badge is not worth an error.
func UnreadNotifications(
	user func(c *web.Ctx) (models.User, error),
	count func(ctx context.Context, userID int64) (int, error),
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			c := web.NewCtx(w, r, web.ServicesFrom(ctx))

			inertia.Share(ctx, "unreadNotifications", func(ctx context.Context) (any, error) {
				u, err := user(c)
				if err != nil {
					return 0, nil //nolint:nilerr // anonymous
				}

				n, err := count(ctx, u.ID)
				if err != nil {
					return 0, nil //nolint:nilerr // badge only
				}

				return n, nil
			})

			next.ServeHTTP(w, r)
		})
	}
}
