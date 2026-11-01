package routes

import (
	"net/http"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/http/problems"
)

// Problems redirects every problem type URL (APP_PROBLEM_BASE_URL + code) to its section of the errors page
// (SDD 05 §6.2). Unknown codes are 404.
func Problems(r *router.Router, _ *vukem.App) error {
	r.Get("/problems/{code}", func(c *web.Ctx) error {
		code := c.Param("code")
		if !problems.Known(code) {
			return errs.ErrNotFound
		}

		return c.RedirectStatus(http.StatusMovedPermanently, problems.DocsPath+"#"+code)
	}).Name("problems.redirect")

	return nil
}
