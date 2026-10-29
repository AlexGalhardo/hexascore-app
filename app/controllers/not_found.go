package controllers

import (
	"net/http"
	"strings"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
)

// NotFound renders the Error page with status 404 for any unmatched GET
// page. API paths keep the default problem response.
func NotFound(c *web.Ctx) error {
	p := c.Request().URL.Path
	if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/app/") {
		return errs.New(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	return inertia.RenderStatus(c, http.StatusNotFound, "Error", inertia.Props{
		"status":  http.StatusNotFound,
		"title":   http.StatusText(http.StatusNotFound),
		"message": http.StatusText(http.StatusNotFound),
	})
}
