package middleware_test

import (
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestVersionHeader(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"APP_VERSION": "v9.9.9"}})
	c := app.Client(t)
	for _, path := range []string{"/", "/up", "/sitemap.xml"} {
		res := c.Get(path)
		if got := res.Header(middleware.VersionHeader); got != "v9.9.9" {
			t.Errorf("%s: %s = %q, want v9.9.9", path, middleware.VersionHeader, got)
		}
	}
}

func TestResolveVersion(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	cases := []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"APP_VERSION": "v0.3.0", "RAILWAY_GIT_COMMIT_SHA": "abcdef123"}, "v0.3.0"},
		{map[string]string{"RAILWAY_GIT_COMMIT_SHA": "abcdef123"}, "abcdef1"},
		{map[string]string{}, "dev"},
	}
	for _, c := range cases {
		if got := middleware.ResolveVersion(env(c.env)); got != c.want {
			t.Errorf("ResolveVersion(%v) = %q, want %q", c.env, got, c.want)
		}
	}
}
