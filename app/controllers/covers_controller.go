package controllers

import (
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/drivers/s3"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// Content-hashed keys never change content; a slug key does when the item's cover changes, so it revalidates
// with the bucket's ETag after a day (owner decision 2026-11-01, item 18).
const (
	coversImmutable = "public, max-age=31536000, immutable"
	coversBySlug    = "public, max-age=86400"
)

// Covers serves GET /covers/{key...} from the S3 bucket (SDD 11 §7: Railway buckets are private, so the
// app streams the objects). Without S3_BUCKET it answers 404 and the embedded demo SVGs are served by
// the static middleware before this route is reached.
func Covers(app *vukem.App) web.HandlerFunc {
	var (
		once sync.Once
		h    http.Handler
	)

	return func(c *web.Ctx) error {
		once.Do(func() {
			var db catalog.Querier
			if d, err := app.DB(); err == nil && d.Dialect().Name() == "postgres" {
				db = d
			}

			h = CoversFrom(BucketDisk(app), db)
		})

		h.ServeHTTP(c.Writer(), c.Request())

		return nil
	}
}

// BucketDisk opens the S3 bucket disk; nil without S3_BUCKET or when the driver fails (logged).
func BucketDisk(app *vukem.App) storage.Disk {
	if app.Source.Get("S3_BUCKET") == "" {
		return nil
	}

	cfg, err := config.Load[s3.Config](app.Source)
	if err == nil {
		var disk storage.Disk
		if disk, err = s3.New(cfg); err == nil {
			return disk
		}
	}

	app.Log.Error("covers: bucket disk", "error", err)

	return nil
}

// legacyCover matches an item or edition cover written before slug keys: <uuid>-<16 hex>[-t].webp.
var legacyCover = regexp.MustCompile(
	`^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})-[0-9a-f]{16}(-t)?\.webp$`,
)

// slugCover matches a slug cover path: {section}/{slug}[-t].webp.
var slugCover = regexp.MustCompile(`^(games|movies|series|books)/([a-z0-9]+(?:-[a-z0-9]+)*)\.webp$`)

// CoversFrom serves covers from disk (nil disk: always 404) for a route with a {key...} wildcard. The key is
// validated exactly like catalog.CoverURL (no leading "/", no ".."), then storage.Handler adds the dotfile,
// backslash and containment checks of the disk. With db (nil skips both steps), a slug path
// (/covers/{section}/{slug}[-t].webp) serves whatever key the item row holds, and an old content-hashed item
// URL answers 301 to its slug path (owner decision 2026-11-01, item 18).
func CoversFrom(disk storage.Disk, db catalog.Querier) http.Handler {
	if disk == nil {
		return http.NotFoundHandler()
	}

	immutable := storage.HandlerWith(disk, nil, storage.Options{CacheControl: coversImmutable})
	bySlug := storage.HandlerWith(disk, nil, storage.Options{CacheControl: coversBySlug})

	serve := func(w http.ResponseWriter, r *http.Request, h http.Handler, key string) {
		p := catalog.CoverURL(key)
		if p == "" || strings.ContainsRune(p, '\\') {
			http.NotFound(w, r)
			return
		}

		r2 := r.Clone(r.Context())
		r2.URL.Path = p
		h.ServeHTTP(w, r2)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if db == nil {
			serve(w, r, coversHandler(key, immutable, bySlug), key)
			return
		}

		if m := slugCover.FindStringSubmatch(key); m != nil {
			stored, err := catalog.ResolveCover(r.Context(), db, sectionType(m[1]), m[2])
			if err != nil || stored == "" {
				coverMiss(w, r, err)
				return
			}

			// The slug URL changes content with the cover, even while its row still holds a hashed key.
			serve(w, r, bySlug, stored)

			return
		}

		if m := legacyCover.FindStringSubmatch(key); m != nil {
			target, err := catalog.CoverPathFor(r.Context(), db, m[1], m[2] != "")
			if err != nil {
				coverMiss(w, r, err)
				return
			}

			// Editions and ambiguous thumbnails keep their content-hashed URL.
			if target != "" {
				// target is "/covers/" plus a section and a catalog slug ([a-z0-9-]), never another host.
				http.Redirect(w, r, target, http.StatusMovedPermanently) //nolint:gosec // see above
				return
			}
		}

		serve(w, r, immutable, key)
	})
}

// coversHandler picks the cache policy of a stored key: content-hashed keys are immutable, slug keys are not.
func coversHandler(key string, immutable, bySlug http.Handler) http.Handler {
	if slugCover.MatchString(strings.TrimPrefix(key, "covers/")) {
		return bySlug
	}

	return immutable
}

func coverMiss(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		slog.ErrorContext(r.Context(), "covers: lookup failed", "path", r.URL.Path, "error", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)

		return
	}

	http.NotFound(w, r)
}

func sectionType(section string) string {
	for typ, s := range catalog.Sections {
		if s == section {
			return typ
		}
	}

	return ""
}
