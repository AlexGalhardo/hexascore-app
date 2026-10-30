package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/appapi"
)

// InstallAppMinVersion sets X-App-Min-Version and answers 426 below
// APP_MIN_SUPPORTED_VERSION. Logout (DELETE /tokens/current) is exempt.
func InstallAppMinVersion(r *router.Router, app *vukem.App) {
	r.UseServer(appapi.MinVersion(app.Source.Get))
}

// AppAuthStack builds the shared /app/v1 guard. extra is ability checks and
// route limits; they run after authentication.
func AppAuthStack(
	app *vukem.App, acc *accounts.Accounts, limiters ratelimit.Factory, extra ...router.Middleware,
) []router.Middleware {
	return appapi.AuthStack(acc, limiters, app.Now, extra...)
}

// AppWriteStack is AppAuthStack plus Idempotency-Key on POST, PATCH, PUT and DELETE.
func AppWriteStack(
	app *vukem.App, acc *accounts.Accounts, limiters ratelimit.Factory, extra ...router.Middleware,
) ([]router.Middleware, error) {
	store, err := app.Cache()
	if err != nil {
		return nil, err
	}

	return append(AppAuthStack(app, acc, limiters, extra...), appapi.Idempotency(store.Store())), nil
}
