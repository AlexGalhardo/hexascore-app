package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/router"
)

// All lists the route files, in order; vukem install:* appends here.
func All() []func(r *router.Router, app *vukem.App) error {
	return []func(r *router.Router, app *vukem.App) error{
		Web,
		Auth,
		Challenge,
		Social,
		TwoFactor,
		Profile,
		Admin,
		Ratings,
		Status,
		Library,
		Lists,
		Follows,
		Notifications,
		Reports,
		APIKeys,
		Problems,
		APIV1,
		APIReference,
		Exports,
		Backups,
		AppAPI,
		AppCatalog,
		AppSearch,
		AppMe,
		AppTelemetry,
		Billing,
		Community,
	}
}
