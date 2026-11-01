package middleware

import (
	"context"
	"net/http"

	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/seo"
)

// ShareConfig configures ShareProps.
type ShareConfig struct {
	AppName string
	// CoversBase is the origin cover keys are served from (COVERS_BASE_URL); empty means the app origin.
	CoversBase string
	SEO        seo.Config
	// User resolves the signed-in user (nil when signed out); it runs only
	// when a page is rendered.
	User func(c *web.Ctx) (any, error)
	// TurnstileSiteKey is the public Turnstile key, empty while Turnstile is off.
	TurnstileSiteKey string
}

// ShareProps adds the props every page receives (usePage().props): appName,
// theme, seo and auth.user (SDD 09 §6.1). Locale, messages and flash come
// from inertia.I18n and inertia.Flash.
func ShareProps(cfg ShareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			c := web.NewCtx(w, r, web.ServicesFrom(ctx))

			inertia.Share(ctx, "appName", cfg.AppName)
			inertia.Share(ctx, "coversBase", cfg.CoversBase)
			inertia.Share(ctx, "turnstileSiteKey", cfg.TurnstileSiteKey)
			inertia.Share(ctx, "seo", cfg.SEO.For(r))
			inertia.Share(ctx, "auth", func(context.Context) (any, error) {
				u, err := cfg.User(c)
				if err != nil {
					u = nil
				}

				return map[string]any{"user": u}, nil
			})

			// The root view reads it as .Page.Props.theme (VF-028: custom view
			// funcs are not request-bound), so only known values pass.
			if ck, err := r.Cookie("theme"); err == nil && (ck.Value == "dark" || ck.Value == "light") {
				inertia.Share(ctx, "theme", ck.Value)
			}

			// The banner is part of the first HTML when this cookie is absent.
			// A client-only open paints after hydration and becomes LCP.
			if ck, err := r.Cookie("consent"); err == nil && (ck.Value == "all" || ck.Value == "essential") {
				inertia.Share(ctx, "consent", ck.Value)
			}

			next.ServeHTTP(w, r)
		})
	}
}
