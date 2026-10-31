package primitives

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
)

func testRecorder(now *time.Time) *Recorder {
	return &Recorder{
		counter: NewCounter(MaxKeys, nil),
		seen:    cache.New(cache.NewMemory(1000), "views:"),
		salt:    func(day string) ([]byte, error) { return []byte("salt-" + day), nil },
		now:     func() time.Time { return *now },
		visited: map[string]int{},
	}
}

func TestCountViewDedupAndDayRollover(t *testing.T) {
	now := time.Date(2026, 10, 31, 23, 59, 0, 0, time.UTC)
	r := testRecorder(&now)

	for i := range 5 {
		if got := r.CountView(t.Context(), "ip:1.2.3.4", "game", "zelda"); got != (i == 0) {
			t.Fatalf("view %d counted=%v", i, got)
		}
	}

	if !r.CountView(t.Context(), "ip:5.6.7.8", "game", "zelda") {
		t.Fatal("another visitor must count")
	}

	now = now.Add(2 * time.Minute) // 00:01 UTC next day: new salt, new claim
	if !r.CountView(t.Context(), "ip:1.2.3.4", "game", "zelda") {
		t.Fatal("same visitor must count again the next UTC day")
	}
}

func TestCountViewCapsDistinctItemsPerVisitor(t *testing.T) {
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	r := testRecorder(&now)

	counted := 0
	for i := range MaxItemsPerVisitor + 10 {
		if r.CountView(t.Context(), "ip:1.1.1.1", "game", string(rune('a'+i%26))+string(rune('a'+i/26))) {
			counted++
		}
	}

	if counted != MaxItemsPerVisitor {
		t.Fatalf("counted %d, want %d", counted, MaxItemsPerVisitor)
	}
}

func TestCountable(t *testing.T) {
	for name, tc := range map[string]struct {
		hdr  map[string]string
		want bool
	}{
		"browser":      {map[string]string{"User-Agent": "Mozilla/5.0 Firefox/130"}, true},
		"prefetch":     {map[string]string{"Sec-Purpose": "prefetch;prerender"}, false},
		"purpose":      {map[string]string{"Purpose": "prefetch"}, false},
		"partial":      {map[string]string{"X-Inertia-Partial-Data": "me"}, false},
		"partial comp": {map[string]string{"X-Inertia-Partial-Component": "Catalog/Show"}, false},
		"googlebot":    {map[string]string{"User-Agent": "Mozilla/5.0 (compatible; Googlebot/2.1)"}, false},
		"lighthouse":   {map[string]string{"User-Agent": "Mozilla/5.0 Chrome-Lighthouse"}, false},
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/games/zelda", nil)
			for k, v := range tc.hdr {
				req.Header.Set(k, v)
			}

			if got := Countable(req); got != tc.want {
				t.Fatalf("Countable = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTarget(t *testing.T) {
	prefixes := map[string]string{"pt-br": "pt-BR", "es": "es"}
	for path, want := range map[string]Key{
		"/games/zelda":             {Type: "game", Ref: "zelda"},
		"/pt-br/books/dune":        {Type: "book", Ref: "dune"},
		"/app/v1/series/abc":       {Type: "series", Ref: "abc"},
		"/games":                   {},
		"/books/dune/editions/1":   {},
		"/people/x":                {},
		"/app/v1/games/genres":     {},
		"/app/v1/games/1/editions": {},
	} {
		got, _ := Target(path, prefixes)
		if got != want {
			t.Errorf("Target(%q) = %+v, want %+v", path, got, want)
		}
	}
}
