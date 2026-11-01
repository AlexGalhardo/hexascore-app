package tmdb_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/tmdb"
)

var _ ingest.Connector = (*tmdb.Connector)(nil)

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile("../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

// fake serves the fixtures under /3 and records the paths and the Authorization header it saw.
type fake struct {
	srv *httptest.Server

	mu    sync.Mutex
	paths []string
	auth  []string
	// status overrides the answer of every request when non-zero.
	status int
}

func newFake(t *testing.T, status int) (*tmdb.Connector, *ingest.Client, *fake) {
	t.Helper()

	f := &fake{status: status}
	files := map[string]string{
		"/3/find/tt0133093": "tmdb_find.json",
		"/3/movie/603":      "tmdb_movie.json",
		"/3/tv/1396":        "tmdb_tv.json",
	}

	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.paths = append(f.paths, r.URL.Path)
		f.auth = append(f.auth, r.Header.Get("Authorization"))
		f.mu.Unlock()

		if f.status != 0 {
			w.WriteHeader(f.status)

			return
		}

		name, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)

			return
		}

		_, _ = w.Write(fixture(t, name))
	}))
	t.Cleanup(f.srv.Close)

	host := strings.TrimPrefix(f.srv.URL, "http://")
	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{host: {PerSecond: 1000, InFlight: 4}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
	})

	c := tmdb.New("token")
	c.APIURL = f.srv.URL + "/3"

	return c, client, f
}

func TestTmdbMovieDetails(t *testing.T) {
	p, err := tmdb.ParseDetails(fixture(t, "tmdb_movie.json"), "movie", "Q83495")
	if err != nil {
		t.Fatal(err)
	}

	if p.ID != 603 || p.URL != "https://www.themoviedb.org/movie/603" {
		t.Errorf("id/url = %d %s", p.ID, p.URL)
	}

	if !strings.HasPrefix(p.Cover, "https://image.tmdb.org/t/p/w780/") {
		t.Errorf("cover = %q", p.Cover)
	}

	if p.Runtime != 136 || p.ReleaseDate != "1999-03-31" {
		t.Errorf("runtime/release = %d %s", p.Runtime, p.ReleaseDate)
	}

	for _, want := range []string{"Lana Wachowski", "Lilly Wachowski"} {
		if !slices.Contains(p.Creators, want) {
			t.Errorf("creators lack %q: %v", want, p.Creators)
		}
	}

	if !slices.Contains(p.Genres, "Science Fiction") {
		t.Errorf("genres = %v", p.Genres)
	}

	if p.Overview == "" {
		t.Error("overview is the input of the synopsis step and must be kept")
	}

	if len(p.Trailers) > tmdb.MaxTrailers {
		t.Errorf("trailers = %v", p.Trailers)
	}

	if p.Seasons != 0 {
		t.Errorf("a movie has no seasons: %d", p.Seasons)
	}
}

func TestTmdbSeriesDetails(t *testing.T) {
	p, err := tmdb.ParseDetails(fixture(t, "tmdb_tv.json"), "tv", "Q1079")
	if err != nil {
		t.Fatal(err)
	}

	if p.URL != "https://www.themoviedb.org/tv/1396" || p.Seasons != 5 || p.Episodes != 62 {
		t.Errorf("payload = %+v", p)
	}

	if !slices.Equal(p.Creators, []string{"Vince Gilligan"}) || p.ReleaseDate != "2008-01-20" {
		t.Errorf("creators/release = %v %s", p.Creators, p.ReleaseDate)
	}
}

func TestTmdbFindPicksTheResultOfTheItemType(t *testing.T) {
	raw := fixture(t, "tmdb_find.json")

	if id, err := tmdb.ParseFind(raw, "movie"); err != nil || id != 603 {
		t.Errorf("movie = %d, %v", id, err)
	}

	if id, err := tmdb.ParseFind(raw, "tv"); err != nil || id != 0 {
		t.Errorf("tv = %d, %v", id, err)
	}
}

func TestTmdbFetchResolvesMissingIDsThroughIMDb(t *testing.T) {
	c, client, f := newFake(t, 0)

	recs, errs := c.Fetch(t.Context(), client, []ingest.Target{
		{ID: "Q83495", Type: tmdb.Ref("movie", "", "tt0133093")},
		{ID: "Q1079", Type: tmdb.Ref("series", "1396", "")},
		{ID: "Q1", Type: tmdb.Ref("movie", "9", "")},
		{ID: "Q2", Type: tmdb.Ref("movie", "", "")},
	})
	if len(errs) != 0 {
		t.Fatalf("a 404 and a target without ids are skips, not errors: %+v", errs)
	}

	if len(recs) != 2 || recs[0].Source != "tmdb" || recs[0].ExternalID != "Q83495" || recs[1].ExternalID != "Q1079" {
		t.Fatalf("records = %+v", recs)
	}

	if !slices.Contains(f.paths, "/3/find/tt0133093") {
		t.Errorf("the IMDb id was not resolved: %v", f.paths)
	}

	for _, a := range f.auth {
		if a != "Bearer token" {
			t.Errorf("authorization = %q", a)
		}
	}
}

func TestTmdbRejectedTokenStopsTheRun(t *testing.T) {
	c, client, _ := newFake(t, http.StatusUnauthorized)

	_, errs := c.Fetch(t.Context(), client, []ingest.Target{{ID: "Q1", Type: tmdb.Ref("movie", "9", "")}})

	if len(errs) != 1 || !errors.Is(errs[0].Err, queue.ErrPermanent) {
		t.Fatalf("a 401 must be permanent: %+v", errs)
	}
}
