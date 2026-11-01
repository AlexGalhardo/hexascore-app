//go:build integration

package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// fakeBucket answers the two S3 calls the exports routes make (ListObjectsV2 and HEAD) from a fixed key set.
func fakeBucket(t *testing.T, keys ...string) *httptest.Server {
	t.Helper()

	have := map[string]bool{}
	for _, k := range keys {
		have["/exports-bucket/"+k] = true
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2":
			var b strings.Builder
			b.WriteString(`<ListBucketResult><IsTruncated>false</IsTruncated>`)

			for _, k := range keys {
				if strings.HasPrefix(k, r.URL.Query().Get("prefix")) {
					b.WriteString("<Contents><Key>" + k + "</Key></Contents>")
				}
			}

			b.WriteString(`</ListBucketResult>`)
			_, _ = w.Write([]byte(b.String()))
		case r.Method == http.MethodHead && have[r.URL.Path]:
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "<Error><Code>NoSuchKey</Code></Error>", http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	return srv
}

func exportsClient(t *testing.T, srv *httptest.Server, email, plan string) *vukemtest.Client {
	t.Helper()

	r := newExportsRig(t, srv)
	token := r.create(r.signIn(email), "ci")

	if plan != "" {
		setPlan(t, r, email, plan)
	}

	return r.app.Client(t).WithToken(token)
}

func newExportsRig(t *testing.T, srv *httptest.Server) *keyRig {
	t.Helper()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{
			"DATABASE_URL": pgtest.Database(t, ""), "OPENAPI_PREFIX": "/api/v1",
			"S3_BUCKET": "exports-bucket", "S3_ENDPOINT": srv.URL, "S3_USE_PATH_STYLE": "true",
			"S3_INSECURE_HTTP": "true", "S3_ACCESS_KEY_ID": "k", "S3_SECRET_ACCESS_KEY": "s",
		},
		AppOptions: []vukem.Option{vukem.WithRateLimiters(func(_ string, _ int, window time.Duration) ratelimit.Limiter {
			return ratelimit.NewMemory(1000, window)
		})},
	})

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	return &keyRig{t: t, app: app, acc: acc, guard: &api.Guard{
		Keys: &api.Keys{DB: acc.DB, Now: app.Now}, User: acc.Users.ByID,
	}}
}

func TestAPIExports(t *testing.T) {
	srv := fakeBucket(t,
		"exports/games_ids_2026_11_01.json.gz",
		"exports/games_ids_2026_11_02.json.gz",
		"exports/series_ids_2026_11_02.json.gz",
		"exports/stray.txt",
	)

	free := exportsClient(t, srv, "exp-free@example.com", "")
	free.GetJSON("/api/v1/exports").AssertForbidden().AssertJSONPath("code", "plan-required")
	free.GetJSON("/api/v1/exports/games/2026-11-02").AssertForbidden().AssertJSONPath("code", "plan-required")

	pro := exportsClient(t, srv, "exp-pro@example.com", "pro")

	t.Run("list", func(*testing.T) {
		pro.GetJSON("/api/v1/exports").AssertOK().AssertMatchesOpenAPI().
			AssertJSONCount("data", 3).
			AssertJSONPath("data.0.type", "games").
			AssertJSONPath("data.0.day", "2026-11-02").
			AssertJSONPath("data.1.type", "series").
			AssertJSONPath("data.2.day", "2026-11-01")
	})

	t.Run("get redirects to a five minute presigned url", func(*testing.T) {
		got := pro.GetJSON("/api/v1/exports/games/2026-11-02").AssertStatus(http.StatusFound)
		loc := got.Header("Location")

		if !strings.HasPrefix(loc, srv.URL+"/exports-bucket/exports/games_ids_2026_11_02.json.gz?") ||
			!strings.Contains(loc, "X-Amz-Expires=300") {
			t.Fatalf("Location = %q, want a presigned URL for the 2026-11-02 games object expiring in 300 s", loc)
		}

		got.AssertHeader("Cache-Control", "no-store")
	})

	t.Run("get refuses what does not exist", func(*testing.T) {
		pro.GetJSON("/api/v1/exports/games/2026-11-03").AssertNotFound()
		pro.GetJSON("/api/v1/exports/series/2026-11-01").AssertNotFound()
		pro.GetJSON("/api/v1/exports/games/yesterday").AssertStatus(http.StatusBadRequest)
		pro.GetJSON("/api/v1/exports/films/2026-11-02").AssertStatus(http.StatusBadRequest)
	})

	t.Run("business is allowed", func(*testing.T) {
		biz := exportsClient(t, srv, "exp-biz@example.com", "business")
		biz.GetJSON("/api/v1/exports/games/2026-11-01").AssertStatus(http.StatusFound)
	})
}
