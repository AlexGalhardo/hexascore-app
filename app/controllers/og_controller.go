package controllers

import (
	"strings"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/seo"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// OGImage serves GET /og/{kind}/{file} (file = "<id>.png"): the share card of a public item, list or
// profile (SDD 09 §7.7, SDD 14 §11 item 12); 404 while the feature flag `social` is off or when an
// anonymous visitor could not see the target.
func OGImage(app *vukem.App, svc *social.Follows) web.HandlerFunc {
	cards := sync.OnceValue(func() *seo.Cards { return seo.NewCards(BucketDisk(app)) })

	return func(c *web.Ctx) error {
		id, ok := strings.CutSuffix(c.Param("file"), ".png")
		if !svc.Enabled(c.Context()) || !ok {
			return NotFound(c)
		}

		db, err := app.DB()
		if err != nil {
			return err
		}

		kind := c.Param("kind")
		key, visible, err := seo.CardCover(c.Context(), db.SQL(), kind, id)
		if err != nil {
			return err
		}
		if !visible {
			return NotFound(c)
		}

		if kind == "profiles" {
			id = strings.ToLower(id)
		}

		out, err := cards().PNG(c.Context(), kind, id, strings.TrimPrefix(catalog.CoverURL(key), "/"))
		if err != nil {
			return err
		}

		h := c.Writer().Header()
		h.Set("Content-Type", "image/png")
		h.Set("Cache-Control", "public, max-age=86400")
		_, err = c.Writer().Write(out)

		return err
	}
}

// ogImage is the share-card path a page passes to <Seo image>, set only while the flag `social` is on.
func ogImage(kind, id string) string { return "/og/" + kind + "/" + id + ".png" }
