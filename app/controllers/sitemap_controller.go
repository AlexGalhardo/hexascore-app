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

// SitemapIndex serves /sitemap.xml (SDD 09 §7.7).
func SitemapIndex(app *vukem.App) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		c.Writer().Header().Set("Content-Type", sitemapType)
		c.Writer().Header().Set("Cache-Control", "public, max-age=3600")

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

		c.Writer().Header().Set("Content-Type", sitemapType)
		c.Writer().Header().Set("Cache-Control", "public, max-age=3600")

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
