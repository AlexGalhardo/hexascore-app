package seo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/seo"
)

var cfg = seo.Config{
	BaseURL: "https://hexascore.app/",
	Locales: []seo.Locale{{Code: "en"}, {Code: "pt-BR", Prefix: "/pt-br"}, {Code: "es", Prefix: "/es"}},
}

func TestFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path, canonical, robots string
	}{
		{"/", "https://hexascore.app/", "index,follow"},
		{"/games/zelda", "https://hexascore.app/games/zelda", "index,follow"},
		{"/pt-br/games/zelda", "https://hexascore.app/pt-br/games/zelda", "index,follow"},
		{"/es/", "https://hexascore.app/es/", "index,follow"},
		{"/es/login", "https://hexascore.app/es/login", "noindex,nofollow"},
		{"/lists/abc", "https://hexascore.app/lists/abc", "noindex,nofollow"},
		{"/listsx", "https://hexascore.app/listsx", "index,follow"},
	}

	for _, tt := range tests {
		m := cfg.For(httptest.NewRequest(http.MethodGet, tt.path+"?q=1", nil))

		if m.Canonical != tt.canonical || m.Robots != tt.robots {
			t.Errorf("%s: canonical %q robots %q", tt.path, m.Canonical, m.Robots)
		}
	}
}

func TestForAlternates(t *testing.T) {
	t.Parallel()

	m := cfg.For(httptest.NewRequest(http.MethodGet, "/pt-br/games/zelda", nil))
	want := map[string]string{
		"en":        "https://hexascore.app/games/zelda",
		"pt-BR":     "https://hexascore.app/pt-br/games/zelda",
		"es":        "https://hexascore.app/es/games/zelda",
		"x-default": "https://hexascore.app/games/zelda",
	}

	if len(m.Alternates) != len(want) {
		t.Fatalf("alternates: %v", m.Alternates)
	}

	for k, v := range want {
		if m.Alternates[k] != v {
			t.Errorf("%s: %q, want %q", k, m.Alternates[k], v)
		}
	}
}
