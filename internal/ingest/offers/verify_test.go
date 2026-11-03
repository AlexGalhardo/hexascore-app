package offers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecide(t *testing.T) {
	t.Parallel()

	const steam = "https://store.steampowered.com/app/292030"

	for _, tc := range []struct {
		name  string
		check Check
		probe Probe
		want  Verdict
	}{
		{
			"ok",
			Check{URL: steam, Status: "unverified"},
			Probe{Status: 200, FinalURL: steam},
			Verdict{Definitive: true, Status: "ok", HTTP: 200, Active: true},
		},
		{
			"ok clears earlier breaks",
			Check{URL: steam, Status: "broken", BrokenRuns: 1},
			Probe{Status: 200, FinalURL: steam},
			Verdict{Definitive: true, Status: "ok", HTTP: 200, Active: true},
		},
		{
			"slug appended is still the same page",
			Check{URL: steam},
			Probe{Status: 200, FinalURL: steam + "/The_Witcher_3/"},
			Verdict{Definitive: true, Status: "ok", HTTP: 200, Active: true},
		},
		{
			"other path",
			Check{URL: steam},
			Probe{Status: 200, FinalURL: "https://store.steampowered.com/"},
			Verdict{Definitive: true, Status: "redirected", HTTP: 200, Active: true},
		},
		{
			"other host",
			Check{URL: steam},
			Probe{Status: 200, FinalURL: "https://example.com/app/292030"},
			Verdict{Definitive: true, Status: "redirected", HTTP: 200, Active: true},
		},
		{
			"first 404",
			Check{URL: steam},
			Probe{Status: 404},
			Verdict{Definitive: true, Status: "broken", HTTP: 404, BrokenRuns: 1, Active: true},
		},
		{
			"second 404 deactivates",
			Check{URL: steam, Status: "broken", BrokenRuns: 1},
			Probe{Status: 404},
			Verdict{Definitive: true, Status: "gone", HTTP: 404, BrokenRuns: 2, Active: false},
		},
		{
			"410 counts like 404",
			Check{URL: steam},
			Probe{Status: 410},
			Verdict{Definitive: true, Status: "broken", HTTP: 410, BrokenRuns: 1, Active: true},
		},
		{"server error is not an answer", Check{URL: steam}, Probe{Status: 503}, Verdict{}},
		{"rate limit is not an answer", Check{URL: steam}, Probe{Status: 429}, Verdict{}},
		{"bot wall is not an answer", Check{URL: steam}, Probe{Status: 403}, Verdict{}},
		{"timeout is not an answer", Check{URL: steam}, Probe{Err: context.DeadlineExceeded}, Verdict{}},
	} {
		if got := Decide(tc.check, tc.probe); got != tc.want {
			t.Errorf("%s: got %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

func TestVerifiableURL(t *testing.T) {
	t.Parallel()

	for url, want := range map[string]bool{
		"https://store.steampowered.com/app/1":              true,
		"https://www.gog.com/en/game/x":                     true,
		"https://play.google.com/store/apps/details?id=a.b": true,
		"https://produto.mercadolivre.com.br/MLB-1":         true,
		"https://www.amazon.com.br/s?k=9780261103573":       false, // Amazon blocks bots
		"https://store.steampowered.com.evil.example/app/1": false, // a look-alike host
		"http://store.steampowered.com/app/1":               false,
		"https://user@store.steampowered.com/app/1":         false,
		"https://example.com/":                              false,
		"https://www.justwatch.com/br/filme/x":              false, // TMDB data is refreshed from the API
	} {
		if got := Verifiable(url); got != want {
			t.Errorf("Verifiable(%q) = %v, want %v", url, got, want)
		}
	}
}

func TestHTTPProberFallsBackToGETWhenHEADIsRefused(t *testing.T) {
	t.Parallel()

	var methods []string

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		_, _ = w.Write(make([]byte, 200<<10)) // more than the 64 KB cap
	}))
	defer srv.Close()

	got := HTTPProber{Client: srv.Client()}.Probe(t.Context(), srv.URL+"/app/1")
	if got.Err != nil || got.Status != 200 {
		t.Fatalf("probe = %+v", got)
	}

	if len(methods) != 2 || methods[0] != http.MethodHead || methods[1] != http.MethodGet {
		t.Fatalf("methods = %v, want HEAD then GET", methods)
	}
}

func TestHTTPProberReportsTransportErrors(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // nothing listens any more

	got := HTTPProber{Client: srv.Client()}.Probe(t.Context(), url)
	if got.Err == nil || errors.Is(got.Err, context.Canceled) {
		t.Fatalf("probe = %+v, want a transport error", got)
	}
}
