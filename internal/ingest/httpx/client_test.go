package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/httpclient"
)

type fakeClock struct {
	mu     sync.Mutex
	now    time.Time
	sleeps []time.Duration
}

func newClock() *fakeClock { return &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if d > 0 {
		c.now = c.now.Add(d)
		c.sleeps = append(c.sleeps, d)
	}

	return ctx.Err()
}

type step struct {
	status int
	header http.Header
	err    error
}

// script replays steps in order and repeats the last one.
type script struct {
	mu    sync.Mutex
	steps []step
	calls int
	seen  []http.Header
	times []time.Time
	clock *fakeClock
}

func (s *script) RoundTrip(req *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.steps[min(s.calls, len(s.steps)-1)]
	s.calls++
	s.seen = append(s.seen, req.Header.Clone())
	s.times = append(s.times, s.clock.Now())

	if st.err != nil {
		return nil, st.err
	}

	h := st.header
	if h == nil {
		h = http.Header{}
	}

	return &http.Response{
		StatusCode: st.status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader("x")),
		Request:    req,
	}, nil
}

func setup(steps ...step) (*Client, *script, *fakeClock) {
	clk := newClock()
	sc := &script{steps: steps, clock: clk}
	c := New(Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]HostLimit{"api.test": {PerSecond: 5}},
		Clock:      clk,
		Base:       sc,
	})

	return c, sc, clk
}

func get(t *testing.T, c *Client, host string) (*http.Response, error) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+host+"/x", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.JSON.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}

	return resp, err
}

func TestUserAgentIdentifiesBot(t *testing.T) {
	c, sc, _ := setup(step{status: 200})

	if _, err := get(t, c, "api.test"); err != nil {
		t.Fatal(err)
	}

	if got := sc.seen[0].Get("User-Agent"); got != "HexaScoreBot/1.0 (+https://hexascore.app/bot)" {
		t.Fatalf("user agent = %q", got)
	}
}

func TestRetryAfterParsing(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"7", 7 * time.Second, true},
		{"0", 0, true},
		{now.Add(90 * time.Second).UTC().Format(http.TimeFormat), 90 * time.Second, true},
		{now.Add(-time.Hour).UTC().Format(http.TimeFormat), 0, true},
		{"soon", 0, false},
		{"-3", 0, false},
		{"", 0, false},
	}

	for _, tc := range cases {
		got, ok := parseRetryAfter(tc.in, now)
		if got != tc.want || ok != tc.ok {
			t.Errorf("parseRetryAfter(%q) = %v, %v; want %v, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestRetries429HonoringRetryAfterThenSucceeds(t *testing.T) {
	c, sc, clk := setup(
		step{status: 429, header: http.Header{"Retry-After": {"30"}}},
		step{status: 200},
	)
	start := clk.Now()

	resp, err := get(t, c, "api.test")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("resp = %v, err = %v", resp, err)
	}

	if sc.calls != 2 {
		t.Fatalf("calls = %d", sc.calls)
	}

	if gap := sc.times[1].Sub(start); gap < 30*time.Second {
		t.Fatalf("second attempt after %v, want >= 30s", gap)
	}
}

func TestGivesUpAfterMaxAttempts(t *testing.T) {
	c, sc, clk := setup(step{status: 503})

	resp, err := get(t, c, "api.test")
	if err != nil || resp.StatusCode != 503 {
		t.Fatalf("resp = %v, err = %v", resp, err)
	}

	if sc.calls != 5 {
		t.Fatalf("calls = %d, want 5", sc.calls)
	}

	// 2^n s backoff between attempts: 2+4+8+16 = 30 s (plus slot spacing).
	if elapsed := clk.Now().Sub(sc.times[0]); elapsed < 30*time.Second {
		t.Fatalf("elapsed %v, want >= 30s", elapsed)
	}
}

func TestRetriesTransportErrors(t *testing.T) {
	c, sc, _ := setup(step{err: errors.New("reset")}, step{status: 200})

	if resp, err := get(t, c, "api.test"); err != nil || resp.StatusCode != 200 || sc.calls != 2 {
		t.Fatalf("resp = %v, err = %v, calls = %d", resp, err, sc.calls)
	}
}

func TestDoesNotRetryClientErrors(t *testing.T) {
	for _, code := range []int{400, 401, 403, 404, 410} {
		c, sc, _ := setup(step{status: code})

		if _, err := get(t, c, "api.test"); err != nil || sc.calls != 1 {
			t.Fatalf("status %d: err = %v, calls = %d", code, err, sc.calls)
		}
	}
}

func TestRetryAfterBeyondCapIsNotWaitedFor(t *testing.T) {
	c, sc, _ := setup(step{status: 429, header: http.Header{"Retry-After": {"301"}}})

	resp, _ := get(t, c, "api.test")
	if resp.StatusCode != 429 || sc.calls != 1 {
		t.Fatalf("status = %d, calls = %d", resp.StatusCode, sc.calls)
	}
}

func TestRateLimitSpacesRequestsPerHost(t *testing.T) {
	c, sc, _ := setup(step{status: 200})

	for range 6 {
		if _, err := get(t, c, "api.test"); err != nil {
			t.Fatal(err)
		}
	}

	for i := 1; i < len(sc.times); i++ {
		if gap := sc.times[i].Sub(sc.times[i-1]); gap < 200*time.Millisecond {
			t.Fatalf("gap %d = %v, want >= 200ms (5/s)", i, gap)
		}
	}
}

func TestRateLimitedResponsePausesTheWholeHost(t *testing.T) {
	clk := newClock()
	sc := &script{steps: []step{{status: 429, header: http.Header{"Retry-After": {"60"}}}, {status: 200}}, clock: clk}
	c := New(
		Config{Hosts: map[string]HostLimit{"a.test": {PerSecond: 5}, "b.test": {PerSecond: 5}}, Clock: clk, Base: sc},
	)
	start := clk.Now()

	if _, err := get(t, c, "a.test"); err != nil {
		t.Fatal(err)
	}

	// The pause belongs to a.test: a request to it right after waits, b.test does not.
	h := c.policyHost("a.test")
	if !h.pausedUntil.After(start.Add(59 * time.Second)) {
		t.Fatalf("pausedUntil = %v", h.pausedUntil)
	}

	if other := c.policyHost("b.test"); !other.pausedUntil.IsZero() {
		t.Fatal("pause leaked to another host")
	}
}

func TestRateHalvesAfterTwentyHits(t *testing.T) {
	c, _, _ := setup(step{status: 429})
	h := c.policyHost("api.test")

	for range 20 {
		h.rateLimited()
	}

	if h.rate != 2.5 {
		t.Fatalf("rate = %v, want 2.5", h.rate)
	}
}

func TestBreakerOpensAndProbes(t *testing.T) {
	c, sc, clk := setup(step{status: 400}) // 4xx first: never a failure
	for range 12 {
		_, _ = get(t, c, "api.test")
	}

	if sc.calls != 12 {
		t.Fatalf("4xx must not open the breaker, calls = %d", sc.calls)
	}

	sc.steps = []step{{status: 500, header: http.Header{"Retry-After": {"0"}}}}

	// 10 consecutive failures open it; later requests fail without a call.
	_, _ = get(t, c, "api.test")
	_, _ = get(t, c, "api.test")

	before := sc.calls
	if _, err := get(t, c, "api.test"); !errors.Is(err, httpclient.ErrCircuitOpen) {
		t.Fatalf("err = %v, want circuit open", err)
	}

	if sc.calls != before {
		t.Fatal("open circuit reached the network")
	}

	// After 5 min one probe goes through; its failure reopens for 10 min.
	clk.now = clk.now.Add(5*time.Minute + time.Second)
	_, _ = get(t, c, "api.test")

	if sc.calls != before+1 {
		t.Fatalf("probe calls = %d, want %d", sc.calls, before+1)
	}

	clk.now = clk.now.Add(6 * time.Minute)
	if _, err := get(t, c, "api.test"); !errors.Is(err, httpclient.ErrCircuitOpen) {
		t.Fatalf("still open after doubled cooldown, err = %v", err)
	}

	// A successful probe closes it.
	clk.now = clk.now.Add(5 * time.Minute)
	sc.steps = []step{{status: 200}}

	if resp, err := get(t, c, "api.test"); err != nil || resp.StatusCode != 200 {
		t.Fatalf("probe: resp = %v, err = %v", resp, err)
	}

	if _, err := get(t, c, "api.test"); err != nil {
		t.Fatalf("closed circuit refused: %v", err)
	}
}

func TestBreakerOpensOnFailureRatio(t *testing.T) {
	var b breakerState

	now := time.Now()
	// alternating keeps consecutive low; 25 failures of 50 reach 50%.
	for i := range 50 {
		b.record(now, i%2 == 1)
	}

	if b.allow(now) {
		t.Fatal("breaker should be open at 50% failures")
	}
}

func TestHostOutsideAllowlistIsRefused(t *testing.T) {
	c, sc, _ := setup(step{status: 200})

	if _, err := get(t, c, "evil.test"); !errors.Is(err, ErrHostNotAllowed) || sc.calls != 0 {
		t.Fatalf("err = %v, calls = %d", err, sc.calls)
	}
}

func (c *Client) policyHost(name string) *host { return c.p.hosts[name] }
