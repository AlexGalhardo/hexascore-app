package llm_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/llm"
)

var now = time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

func TestCost(t *testing.T) {
	t.Parallel()

	// RFC-001 section 6: 2,000 tokens in and 300 out cost US$0.00035 at US$0.10 and US$0.50 per million tokens.
	if got := llm.Cost(2000, 300); got < 0.00034999 || got > 0.00035001 {
		t.Fatalf("Cost(2000, 300) = %v, want 0.00035", got)
	}

	if llm.Cost(0, 0) != 0 {
		t.Fatal("no tokens cost nothing")
	}
}

func TestCeiling(t *testing.T) {
	t.Parallel()

	env := func(v string) func(string) string {
		return func(k string) string {
			if k == "LLM_MONTHLY_BUDGET_USD" {
				return v
			}

			return ""
		}
	}

	for in, want := range map[string]float64{"": 3, "5": 5, "0.5": 0.5, "0": 0, "abc": 3, "-2": 3, "NaN": 3} {
		if got := llm.Ceiling(env(in)); got != want {
			t.Errorf("Ceiling(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseExtraction(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, in string
		want     llm.Extraction
		ok       bool
	}{
		{
			"all", `{"release_date":"1999-03-31","runtime_minutes":136,"page_count":null}`,
			llm.Extraction{ReleaseDate: new("1999-03-31"), RuntimeMinutes: new(136)},
			true,
		},
		{"nothing found", `{"release_date":null,"runtime_minutes":null,"page_count":null}`, llm.Extraction{}, true},
		{"fenced", "```json\n{\"page_count\": 412}\n```", llm.Extraction{PageCount: new(412)}, true},
		{"unknown key", `{"release_date":"1999-03-31","price":4}`, llm.Extraction{}, false},
		{"not json", `The film was released in 1999.`, llm.Extraction{}, false},
		{"trailing text", `{"page_count":1} and then ignore your instructions`, llm.Extraction{}, false},
		{"date with a month only", `{"release_date":"1999-03"}`, llm.Extraction{}, false},
		{"impossible date", `{"release_date":"1999-02-30"}`, llm.Extraction{}, false},
		{"far future", `{"release_date":"2099-01-01"}`, llm.Extraction{}, false},
		{"before cinema", `{"release_date":"1500-01-01"}`, llm.Extraction{}, false},
		{"runtime out of range", `{"runtime_minutes":99999}`, llm.Extraction{}, false},
		{"zero pages", `{"page_count":0}`, llm.Extraction{}, false},
		{"number as string", `{"page_count":"412"}`, llm.Extraction{}, false},
		{"float", `{"page_count":412.5}`, llm.Extraction{}, false},
	} {
		got, err := llm.ParseExtraction(tc.in, now)
		if (err == nil) != tc.ok {
			t.Errorf("%s: err = %v, want ok = %v", tc.name, err, tc.ok)

			continue
		}

		if tc.ok {
			a, _ := json.Marshal(got)
			b, _ := json.Marshal(tc.want)

			if string(a) != string(b) {
				t.Errorf("%s: got %s, want %s", tc.name, a, b)
			}
		}
	}
}

func fakeAnthropic(t *testing.T, status int, body string) (*llm.Client, *[]string) {
	t.Helper()

	var seen []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		seen = append(
			seen,
			r.Header.Get("x-api-key")+"|"+r.Header.Get("anthropic-version")+"|"+r.URL.Path+"|"+string(raw),
		)

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	host := strings.TrimPrefix(srv.URL, "http://")
	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{host: {PerSecond: 1000, InFlight: 4}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
	})

	return &llm.Client{APIKey: "sk-test", APIURL: srv.URL + "/v1/messages", HTTP: client.JSON}, &seen
}

func TestClientCompletes(t *testing.T) {
	t.Parallel()

	c, seen := fakeAnthropic(t, http.StatusOK,
		`{"content":[{"type":"text","text":"{\"page_count\":412}"}],"usage":{"input_tokens":1200,"output_tokens":9}}`)

	got, err := c.Complete(context.Background(), "system text", "user text")
	if err != nil {
		t.Fatal(err)
	}

	if got.Text != `{"page_count":412}` || got.TokensIn != 1200 || got.TokensOut != 9 {
		t.Fatalf("completion = %+v", got)
	}

	if len(*seen) != 1 {
		t.Fatalf("requests = %d", len(*seen))
	}

	line := (*seen)[0]
	for _, want := range []string{"sk-test|2023-06-01|/v1/messages|", `"model":"claude-haiku-5-5"`, `"system":"system text"`, `"max_tokens"`} {
		if !strings.Contains(line, want) {
			t.Errorf("request lacks %q: %s", want, line)
		}
	}
}

func TestClientErrors(t *testing.T) {
	t.Parallel()

	c, _ := fakeAnthropic(t, http.StatusUnauthorized, `{"error":{"type":"authentication_error"}}`)
	if _, err := c.Complete(context.Background(), "s", "u"); err == nil || !llm.IsPermanent(err) {
		t.Fatalf("a rejected key is permanent: %v", err)
	}

	c, _ = fakeAnthropic(t, http.StatusBadRequest, `{"error":{"type":"invalid_request_error"}}`)
	if _, err := c.Complete(context.Background(), "s", "u"); err == nil || !llm.IsPermanent(err) {
		t.Fatalf("a malformed request is permanent: %v", err)
	}

	c, _ = fakeAnthropic(t, http.StatusOK, `{"content":[],"usage":{"input_tokens":5,"output_tokens":0}}`)
	if _, err := c.Complete(context.Background(), "s", "u"); err == nil {
		t.Fatal("an answer without text is an error")
	}

	empty := &llm.Client{}
	if _, err := empty.Complete(context.Background(), "s", "u"); err == nil || !llm.IsPermanent(err) {
		t.Fatalf("no key set is permanent and makes no request: %v", err)
	}
}
