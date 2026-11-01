package steam_test

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/steam"
)

var _ ingest.Connector = (*steam.Connector)(nil)

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile("../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func TestSteamAppdetails(t *testing.T) {
	p, err := steam.ParseAppdetails(fixture(t, "steam_appdetails.json"), "292030", "Q4267401")
	if err != nil || p == nil {
		t.Fatalf("ParseAppdetails = %v, %v", p, err)
	}

	if !slices.Equal(p.Platforms, []string{"Windows"}) {
		t.Errorf("platforms = %v", p.Platforms)
	}

	if !strings.HasPrefix(p.HeaderImage, "https://") {
		t.Errorf("header_image = %q", p.HeaderImage)
	}

	if p.StoreURL != "https://store.steampowered.com/app/292030/" {
		t.Errorf("store_url = %q", p.StoreURL)
	}

	failed, err := steam.ParseAppdetails(fixture(t, "steam_appdetails_fail.json"), "1", "Q1")
	if err != nil || failed != nil {
		t.Errorf("an unsuccessful envelope must give no payload: %v, %v", failed, err)
	}
}

func TestEnrichmentSkipsFailedTargets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("appids") == "10" {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		_, _ = w.Write(fixture(t, "steam_appdetails.json"))
	}))
	t.Cleanup(srv.Close)

	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{strings.TrimPrefix(srv.URL, "http://"): {PerSecond: 1000, InFlight: 1}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
	})

	c := &steam.Connector{APIURL: srv.URL}

	records, errs := c.Fetch(t.Context(), client, []ingest.Target{{ID: "Q1", Type: "10"}, {ID: "Q2", Type: "292030"}})
	if len(records) != 1 || records[0].ExternalID != "Q2" || records[0].Source != "steam" {
		t.Errorf("records = %+v, want only Q2", records)
	}

	if len(errs) != 1 || errs[0].Target.ID != "Q1" {
		t.Errorf("errs = %v, want the failed Q1 kept for the dead-letter table", errs)
	}
}
