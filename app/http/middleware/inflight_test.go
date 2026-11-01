package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

func TestInflightCountsRequestsAndTimesItemPages(t *testing.T) {
	p := ingest.NewPressure(time.Now)

	var during int

	h := Inflight(p)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { during = p.InFlight() }))

	for _, path := range []string{"/movies/dune", "/pt-BR/games/celeste", "/search", "/movies/dune/changes"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	if during != 1 {
		t.Errorf("in flight during a request = %d, want 1", during)
	}

	if p.InFlight() != 0 {
		t.Errorf("in flight after the requests = %d, want 0", p.InFlight())
	}

	for _, c := range []struct {
		path string
		want bool
	}{{"/movies/dune", true}, {"/pt-BR/games/celeste", true}, {"/books/x/", true}, {"/movies/dune/changes", false}, {"/people/ana", false}} {
		if got := itemPath.MatchString(c.path); got != c.want {
			t.Errorf("itemPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
