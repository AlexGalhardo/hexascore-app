package ingest_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/AlexGalhardo/vukem/httpclient"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
)

func TestSecurity_A01_IngestSSRF(t *testing.T) {
	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
	t.Cleanup(srv.Close)

	host := srv.Listener.Addr().String()

	cases := []struct {
		name  string
		hosts map[string]httpx.HostLimit
		want  error
	}{
		{"private ip even when allowlisted", map[string]httpx.HostLimit{host: {}}, httpclient.ErrBlockedAddress},
		{"host outside the allowlist", map[string]httpx.HostLimit{"query.wikidata.org": {}}, httpx.ErrHostNotAllowed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := httpx.New(httpx.Config{ContactURL: "https://hexascore.app/bot", Hosts: tc.hosts})

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := c.JSON.Do(req)
			if resp != nil {
				_ = resp.Body.Close()
			}

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}

			if hits.Load() != 0 {
				t.Fatal("request reached the server")
			}
		})
	}
}
