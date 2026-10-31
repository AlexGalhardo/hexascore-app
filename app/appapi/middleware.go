package appapi

import (
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
)

// AuthStack is the /app/v1 guard (SDD 05 §6.4): JSON, an IP limit, the hx_app_
// token guard, NotBanned, Verified, then 120 reads or 30 writes per minute
// per token. extra runs after that guard and before Idempotency when the
// caller appends it. Not plan-metered.
func AuthStack(
	acc *accounts.Accounts,
	limiters ratelimit.Factory,
	now func() time.Time,
	extra ...router.Middleware,
) []router.Middleware {
	reads := ratelimit.Middleware(limits.AppRead.From(limiters), ratelimit.ByToken)
	writes := ratelimit.Middleware(limits.AppWrite.From(limiters), ratelimit.ByToken)

	out := []router.Middleware{
		middleware.JSONOnly(),
		ratelimit.Middleware(limits.AppIP.From(limiters), ratelimit.ByIP),
		auth.Required(acc.API),
		appmiddleware.NotBanned(acc.API.User, acc.ProblemBase, now),
		auth.Verified(acc.API),
		tokenBudget(reads, writes),
	}

	return append(out, extra...)
}

// Idempotency replays a completed write and answers 409 while it is in
// flight, 422 when the same key is reused for a different body. It must run
// after auth.Required so the key is scoped to the caller.
func Idempotency(store cache.Store) router.Middleware {
	return api.Idempotency(store, api.IdempotencyOptions{
		Methods: []string{http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete},
	})
}

func tokenBudget(read, write router.Middleware) router.Middleware {
	return func(next http.Handler) http.Handler {
		readNext := read(next)
		writeNext := write(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				readNext.ServeHTTP(w, r)
				return
			}

			writeNext.ServeHTTP(w, r)
		})
	}
}
