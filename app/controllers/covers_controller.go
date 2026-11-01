package controllers

import (
	"net/http"
	"strings"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/drivers/s3"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

const coversCacheControl = "public, max-age=31536000, immutable"

// Covers serves GET /covers/{key...} from the S3 bucket (SDD 11 §7: Railway buckets are private, so the
// app streams the objects). Without S3_BUCKET it answers 404 and the embedded demo SVGs are served by
// the static middleware before this route is reached.
func Covers(app *vukem.App) web.HandlerFunc {
	var (
		once sync.Once
		h    http.Handler
	)

	return func(c *web.Ctx) error {
		once.Do(func() { h = CoversFrom(BucketDisk(app)) })

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

// CoversFrom serves covers from disk (nil disk: always 404) for a route with a {key...} wildcard. The
// key is validated exactly like catalog.CoverURL (no leading "/", no ".."), then storage.Handler adds
// the dotfile, backslash and containment checks of the disk.
func CoversFrom(disk storage.Disk) http.Handler {
	if disk == nil {
		return http.NotFoundHandler()
	}

	serve := storage.HandlerWith(disk, nil, storage.Options{CacheControl: coversCacheControl})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := catalog.CoverURL(r.PathValue("key"))
		if p == "" || strings.ContainsRune(p, '\\') {
			http.NotFound(w, r)
			return
		}

		r2 := r.Clone(r.Context())
		r2.URL.Path = p
		serve.ServeHTTP(w, r2)
	})
}
