package igdb_test

import (
	"errors"
	"fmt"
	"io"
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
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/igdb"
)

var _ ingest.Connector = (*igdb.Connector)(nil)

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile("../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func TestIgdbGames(t *testing.T) {
	raw := fixture(t, "igdb_games.json")

	none, err := igdb.ParseGames(raw, map[string]string{"some-other-game": "Q1"})
	if err != nil || len(none) != 0 {
		t.Fatalf("games outside the batch must be ignored: %v %v", none, err)
	}

	got, err := igdb.ParseGames(raw, map[string]string{"the-witcher-3-wild-hunt": "Q4267401"})
	if err != nil || len(got) != 1 {
		t.Fatalf("ParseGames = %v, %v", got, err)
	}

	p := got[0]
	if p.Qid != "Q4267401" || p.URL != "https://www.igdb.com/games/the-witcher-3-wild-hunt" {
		t.Errorf("qid/url = %s %s", p.Qid, p.URL)
	}

	if p.Cover != "https://images.igdb.com/igdb/image/upload/t_cover_big_2x/coaarl.jpg" {
		t.Errorf("cover = %q", p.Cover)
	}

	if p.ReleaseDate != "2015-05-19" {
		t.Errorf("release_date = %q", p.ReleaseDate)
	}

	if !slices.Equal(p.Creators, []string{"CD Projekt RED"}) {
		t.Errorf("creators = %v", p.Creators)
	}

	if !slices.Contains(p.Platforms, "PlayStation 4") {
		t.Errorf("platforms = %v", p.Platforms)
	}

	if !slices.Equal(p.Trailers, []string{"yowv6_rspoM", "xQGam9OHSUo", "QrwGXAcE6ZA"}) {
		t.Errorf("trailers = %v", p.Trailers)
	}

	if p.Overview == "" {
		t.Error("the summary is the input of the synopsis step and must be kept")
	}
}

func TestIgdbReleaseDatesBefore1970(t *testing.T) {
	got, err := igdb.ParseGames(
		[]byte(`[{"id": 1, "slug": "spacewar", "first_release_date": -244252800}]`),
		map[string]string{"spacewar": "Q1"},
	)
	if err != nil || len(got) != 1 {
		t.Fatalf("ParseGames = %v, %v", got, err)
	}

	if got[0].ReleaseDate != "1962-04-06" {
		t.Errorf("release_date = %q, want 1962-04-06", got[0].ReleaseDate)
	}
}

// fake serves the Twitch token endpoint and the IGDB games endpoint.
type fake struct {
	srv *httptest.Server

	mu     sync.Mutex
	grants []string
	bodies []string
	heads  []string
	status int
	// expiresIn is the token lifetime the fake reports (default 5000 s); reject401 answers 401 to that many
	// games calls first.
	expiresIn int
	reject401 int
}

func newFake(t *testing.T, status int) (*ingest.Client, *fake) {
	t.Helper()

	f := &fake{status: status}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		if r.URL.RawQuery != "" {
			t.Errorf("token request carries a query string: %q", r.URL.RawQuery)
		}

		f.grants = append(f.grants, r.PostFormValue("grant_type")+"|"+r.PostFormValue("client_id"))
		f.mu.Unlock()

		exp := f.expiresIn
		if exp == 0 {
			exp = 5000
		}
		_, _ = fmt.Fprintf(w, `{"access_token": "abc", "expires_in": %d, "token_type": "bearer"}`, exp)
	})
	mux.HandleFunc("/v4/games", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		f.mu.Lock()
		f.bodies = append(f.bodies, string(body))
		f.heads = append(f.heads, r.Header.Get("Client-ID")+"|"+r.Header.Get("Authorization"))
		reject := f.reject401 > 0
		if reject {
			f.reject401--
		}
		f.mu.Unlock()

		if reject {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if f.status != 0 {
			w.WriteHeader(f.status)

			return
		}

		_, _ = w.Write(fixture(t, "igdb_games.json"))
	})

	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)

	host := strings.TrimPrefix(f.srv.URL, "http://")
	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{host: {PerSecond: 1000, InFlight: 4}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
	})

	return client, f
}

func connector(f *fake, id, secret string) *igdb.Connector {
	c := igdb.New(id, secret)
	c.TokenURL = f.srv.URL + "/oauth2/token"
	c.APIURL = f.srv.URL + "/v4"

	return c
}

func TestIgdbFetchAuthenticatesAndDropsUnsafeSlugs(t *testing.T) {
	client, f := newFake(t, 0)
	c := connector(f, "id", "secret")

	records, errs := c.Fetch(t.Context(), client, []ingest.Target{
		{ID: "Q4267401", Type: "the-witcher-3-wild-hunt"},
		{ID: "Q2", Type: `x"; fields *;`},
	})
	if len(errs) != 0 {
		t.Fatalf("errs = %v", errs)
	}

	if len(records) != 1 || records[0].ExternalID != "Q4267401" || records[0].Source != "igdb" {
		t.Fatalf("records = %+v", records)
	}

	if !slices.Equal(f.grants, []string{"client_credentials|id"}) {
		t.Errorf("token requests = %v", f.grants)
	}

	want := "fields " + igdb.Fields + `; where slug = ("the-witcher-3-wild-hunt"); limit 1;`
	if !slices.Equal(f.bodies, []string{want}) {
		t.Errorf("bodies = %q, want %q", f.bodies, want)
	}

	if !slices.Equal(f.heads, []string{"id|Bearer abc"}) {
		t.Errorf("headers = %v", f.heads)
	}
}

func TestIgdbMissingCredentialsOrRejectedTokenStopTheRun(t *testing.T) {
	client, f := newFake(t, http.StatusUnauthorized)

	for name, c := range map[string]*igdb.Connector{
		"no credentials": connector(f, "", ""),
		"rejected token": connector(f, "id", "secret"),
	} {
		_, errs := c.Fetch(t.Context(), client, []ingest.Target{{ID: "Q4267401", Type: "the-witcher-3-wild-hunt"}})
		if len(errs) != 1 || !errors.Is(errs[0].Err, queue.ErrPermanent) {
			t.Errorf("%s: errs = %v, want one permanent error", name, errs)
		}
	}
}

func TestToken_FetchedOncePerExpiry(t *testing.T) {
	target := []ingest.Target{{ID: "Q4267401", Type: "the-witcher-3-wild-hunt"}}

	client, f := newFake(t, 0)
	c := connector(f, "id", "secret")
	for range 3 {
		if _, errs := c.Fetch(t.Context(), client, target); len(errs) != 0 {
			t.Fatalf("errs = %v", errs)
		}
	}
	if len(f.grants) != 1 {
		t.Errorf("token requests over 3 fetches = %d, want 1 (cached until expires_in)", len(f.grants))
	}

	// A token that expires within the renewal margin is never reused.
	client, f = newFake(t, 0)
	f.expiresIn = 60
	c = connector(f, "id", "secret")
	for range 2 {
		if _, errs := c.Fetch(t.Context(), client, target); len(errs) != 0 {
			t.Fatalf("errs = %v", errs)
		}
	}
	if len(f.grants) != 2 {
		t.Errorf("token requests with a short-lived token = %d, want 2", len(f.grants))
	}
}

func TestToken_RefreshedOn401(t *testing.T) {
	client, f := newFake(t, 0)
	f.reject401 = 1
	c := connector(f, "id", "secret")

	records, errs := c.Fetch(t.Context(), client, []ingest.Target{{ID: "Q4267401", Type: "the-witcher-3-wild-hunt"}})
	if len(errs) != 0 || len(records) != 1 {
		t.Fatalf("records = %d, errs = %v; want the chunk retried with a new token", len(records), errs)
	}
	if len(f.grants) != 2 {
		t.Errorf("token requests = %d, want 2 (initial + refresh after 401)", len(f.grants))
	}
}
