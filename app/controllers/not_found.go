package controllers

import (
	"net/http"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"
)

// NotFound answers a missing page (unknown route, item, entity or document) with 404: the error page, localised
// and linking home, for HTML; problem+json for the APIs (decision D-41, 2026-11-02, replacing the 301 to the home
// page of 2026-11-01: search engines read that as a soft 404 and users lost the link they typed). Merged items and
// entities keep their 301 through catalog_redirects before a handler reaches this.
func NotFound(_ *web.Ctx) error {
	return errs.New(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

// pageError turns a 404 from a service into NotFound; any other error passes through unchanged.
func pageError(c *web.Ctx, err error) error {
	if errs.Status(err) == http.StatusNotFound {
		return NotFound(c)
	}

	return err
}
