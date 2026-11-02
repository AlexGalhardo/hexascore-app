package controllers

import (
	"errors"
	"strings"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

const sitemapType = "application/xml; charset=utf-8"

// sitemapHeaders sets the XML headers and drops the CSP: a sitemap runs nothing, and the policy blocks the styles of
// Chrome's XML tree viewer, so the file showed as raw text (owner 2026-11-02).
func sitemapHeaders(c *web.Ctx) {
	h := c.Writer().Header()
	h.Set("Content-Type", sitemapType)
	h.Set("Cache-Control", "public, max-age=3600")
	h.Del("Content-Security-Policy")
}

// SitemapIndex serves /sitemap.xml (SDD 09 §7.7).
func SitemapIndex(app *vukem.App) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		sitemapHeaders(c)

		return catalog.WriteSitemapIndex(c.Context(), db, c.Writer(), origin(app))
	}
}

// Sitemap serves /sitemaps/{file}: static.xml and the per-type chunks.
func Sitemap(app *vukem.App) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		sitemapHeaders(c)

		err = catalog.WriteSitemap(c.Context(), db, c.Writer(), origin(app), c.Param("file"))
		if errors.Is(err, catalog.ErrNoSitemap) {
			c.Writer().Header().Del("Cache-Control")
			c.Writer().Header().Del("Content-Type")

			return errs.ErrNotFound
		}

		return err
	}
}

func origin(app *vukem.App) string { return strings.TrimRight(app.Config.URL, "/") }
