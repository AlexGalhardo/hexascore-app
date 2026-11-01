package controllers

import (
	"net/http"
	"path"
	"strings"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/web"
)

// realMissPrefixes keep a real 404 on a miss: the APIs (and /v1, the API host's alias) answer problem+json, and an
// asset or feed miss must not turn into a redirect a cache or an image tag would follow.
var realMissPrefixes = []string{
	"/api", "/app", "/v1", "/build", "/covers", "/og", "/sitemaps", "/.well-known", "/search/suggest", "/up",
}

// pageMiss reports whether a miss on c is an HTML page: a GET or HEAD that does not ask for JSON, outside
// realMissPrefixes, whose last segment has no file extension. The path is the routed one, the locale prefix
// already stripped.
func pageMiss(c *web.Ctx) bool {
	r := c.Request()
	if r.Method != http.MethodGet && r.Method != http.MethodHead || c.WantsJSON() {
		return false
	}

	p := r.URL.Path
	if path.Ext(p) != "" {
		return false
	}

	for _, prefix := range realMissPrefixes {
		if p == prefix || strings.HasPrefix(p, prefix+"/") {
			return false
		}
	}

	return true
}

// NotFound answers a missing page (unknown route, item, entity or document) with 301 to the home page of the
// request's locale (owner decision 2026-11-01, item 14). API and asset paths, and non-GET requests, keep the
// 404: problem+json for the APIs, the error page otherwise.
func NotFound(c *web.Ctx) error {
	if pageMiss(c) {
		return c.RedirectStatus(http.StatusMovedPermanently, i18n.PathPrefix(c.Context())+"/")
	}

	return errs.New(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

// pageError turns a 404 from a service into NotFound; any other error passes through unchanged.
func pageError(c *web.Ctx, err error) error {
	if errs.Status(err) == http.StatusNotFound {
		return NotFound(c)
	}

	return err
}
