package primitives

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"go.opentelemetry.io/otel"
)

type dbFunc func() (*database.DB, error)

// upsert adds one counter row to item_activity_daily. The item is resolved here, so a page view needs no lookup: ref
// is a slug or an id, and an empty type (search clicks) matches any. A key that is not a public item inserts nothing.
// The id branch casts only a UUID-shaped ref, so both branches stay index lookups; `id::text = ref` scanned every item.
const upsert = `INSERT INTO item_activity_daily (item_id, day, item_type, views, search_clicks)
SELECT i.id, $3::date, i.type, $4::int, $5::int FROM items i
WHERE i.visibility = 'public' AND ($1::text = '' OR i.type = $1::text)
AND ((i.type = $1::text AND i.slug = $2::text)
	OR i.id = CASE WHEN $2::text ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN $2::uuid END)
LIMIT 1
ON CONFLICT (item_id, day) DO UPDATE SET
	views = item_activity_daily.views + EXCLUDED.views,
	search_clicks = item_activity_daily.search_clicks + EXCLUDED.search_clicks`

// Flush writes the buffered counts in one transaction. On failure the counts go back into the buffer and join the
// next flush (SDD 08 section 8), so a retry never counts anything twice.
func (r *Recorder) Flush(ctx context.Context) error {
	batch := r.counter.Swap()
	if len(batch) == 0 {
		return nil
	}

	db, err := r.db()
	if err == nil {
		err = db.Tx(ctx, func(tx *database.Tx) error {
			for k, t := range batch {
				if _, err := tx.ExecContext(ctx, upsert, k.Type, k.Ref, k.Day, t.Views, t.Clicks); err != nil {
					return err
				}
			}

			return nil
		})
	}

	if err != nil {
		r.counter.Merge(batch)

		return fmt.Errorf("trending flush: %w", err)
	}

	return nil
}

// Middleware counts 200 GETs of item pages and /app/v1 item reads. localePrefixes are the first path segments the
// app serves a locale under. Mount it outside the render cache with UseServer, so pages served from the cache count
// too.
func (r *Recorder) Middleware(sessionCookie string, localePrefixes map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			k, app := Target(req.URL.Path, localePrefixes)
			if req.Method != http.MethodGet || k.Ref == "" || !Countable(req) {
				next.ServeHTTP(w, req)

				return
			}

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, req)

			if sw.status == http.StatusOK {
				r.CountView(req.Context(), visitor(req, sessionCookie, app), k.Type, k.Ref)
			}
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// visitor is the dedup identity (SDD 08 FR-08-1): the access token for /app/v1, else the session cookie, else the
// IP grouped by the IPv6 prefix. It only ever reaches memory as an HMAC input.
func visitor(req *http.Request, sessionCookie string, app bool) string {
	if app {
		return ratelimit.ByToken(req)
	}

	for _, name := range []string{"__Host-" + sessionCookie, sessionCookie} {
		if c, err := req.Cookie(name); err == nil && c.Value != "" {
			return "session:" + c.Value
		}
	}

	return ratelimit.ByIP(req)
}

var recorders sync.Map

// For returns the Recorder of app, creating it on first use and hooking its final flush to app shutdown.
func For(app *vukem.App) *Recorder {
	if r, ok := recorders.Load(app); ok {
		if rec, ok := r.(*Recorder); ok {
			return rec
		}
	}

	dropped, _ := otel.Meter("hexascore").Int64Counter("trending.counter.dropped")

	var (
		mu      sync.Mutex
		saltDay string
		saltKey []byte
	)

	r := &Recorder{
		seen: cache.New(cache.NewMemory(dedupEntries), "views:"),
		now:  app.Now,
		db:   app.DB,
		salt: func(day string) ([]byte, error) {
			mu.Lock()
			defer mu.Unlock()

			if saltDay != day {
				k, err := app.DeriveKey("trending.views."+day, 32)
				if err != nil {
					return nil, err
				}

				saltDay, saltKey = day, k
			}

			return saltKey, nil
		},
		visited: map[string]int{},
		counter: NewCounter(MaxKeys, func(n int) {
			if dropped != nil {
				dropped.Add(context.Background(), int64(n))
			}
		}),
	}

	if actual, loaded := recorders.LoadOrStore(app, r); loaded {
		if rec, ok := actual.(*Recorder); ok {
			return rec
		}
	}

	app.OnShutdown(func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()

		return r.Flush(ctx)
	})

	return r
}
