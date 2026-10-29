package controllers

import (
	"fmt"
	"strings"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/seo"
)

// Robots serves /robots.txt (SDD 09 §7.7). Only production is indexable, so
// previews and local builds never compete with it in search results.
func Robots(app *vukem.App) web.HandlerFunc {
	var b strings.Builder

	b.WriteString("User-agent: *\n")

	if app.Config.IsProduction() {
		b.WriteString("Allow: /\n")

		for _, rule := range seo.Disallowed() {
			fmt.Fprintf(&b, "Disallow: %s\n", rule)
		}
	} else {
		b.WriteString("Disallow: /\n")
	}

	fmt.Fprintf(&b, "\nSitemap: %s/sitemap.xml\n", origin(app))
	body := b.String()

	return func(c *web.Ctx) error {
		c.Writer().Header().Set("Content-Type", "text/plain; charset=utf-8")
		c.Writer().Header().Set("Cache-Control", "public, max-age=3600")
		_, err := c.Writer().Write([]byte(body))

		return err
	}
}
