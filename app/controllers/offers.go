package controllers

import (
	"errors"
	"net/http"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/offers"
)

// OfferGo serves GET /go/{id}, the only way out of an offer link (RFC-003 section 4). The target is read from the
// offers table by id and nothing else: no parameter names a URL, so the route cannot be used as an open redirect. The
// affiliate tag is added here, never stored. The click is counted by day, offer and country with no user id and no IP,
// and the response is not cached, carries no referrer and is kept out of search results.
func OfferGo(app *vukem.App) web.HandlerFunc {
	tag := offers.EnvTagger(app.Source.Get)

	return func(c *web.Ctx) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		now := app.Now().UTC()

		link, err := offers.Resolve(c.Context(), db, c.Param("id"), now)
		if errors.Is(err, offers.ErrNotFound) {
			return errs.ErrNotFound
		}

		if err != nil {
			return err
		}

		// A counter that fails must not strand the visitor: the redirect still goes out.
		if err := offers.Click(c.Context(), db, c.Param("id"), link.Country, now); err != nil {
			app.Log.WarnContext(c.Context(), "offer click not counted", "error", err)
		}

		c.SetHeader("Cache-Control", "no-store")
		c.SetHeader("Referrer-Policy", "no-referrer")
		c.SetHeader("X-Robots-Tag", "noindex, nofollow")

		return c.RedirectStatus(http.StatusFound, link.Final(tag))
	}
}
