// Package httpx is the outbound client of the ingestion pipeline (SDD 06
// §7.2). It builds on vukem's httpclient.Safe, which supplies the SSRF dial
// check, redirect limit and body cap. The vukem v0.34.0 Throttle, Retry and
// Breaker transports are not used because they run on the wall clock (no
// fake clock for NFR-06-5), cannot pause a whole host on Retry-After, halve
// a rate, retry 500, or open on a failure ratio with a doubling cooldown
// (VF-020 stays partly open).
package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/httpclient"
)

// ErrHostNotAllowed is returned before any connection for a host outside
// the allowlist.
var ErrHostNotAllowed = errors.New("httpx: host is not allowlisted")

const (
	maxAttempts   = 5
	maxBackoff    = 60 * time.Second
	maxRetryAfter = 300 * time.Second
	timeout       = 60 * time.Second
	// MaxJSONBytes caps JSON and text responses.
	MaxJSONBytes = 20 << 20
	// MaxImageBytes caps image downloads.
	MaxImageBytes = 15 << 20
)

// Config configures a Client. Hosts is both the allowlist and the rate
// table, keyed by URL host (with port when not default).
type Config struct {
	// ContactURL fills the User-Agent (INGEST_CONTACT_URL).
	ContactURL string
	Hosts      map[string]HostLimit
	// Allow lists extra reachable prefixes; tests only.
	Allow []netip.Prefix
	Clock Clock
	// Base replaces the Safe transport; tests only.
	Base http.RoundTripper
	// Timeout bounds one attempt; zero means 60 s. Tests only.
	Timeout time.Duration
}

// Client holds one client per body cap, sharing throttle and breaker state.
type Client struct {
	JSON  *http.Client
	Image *http.Client

	p *policy
}

// UserAgent returns the bot User-Agent.
func UserAgent(contactURL string) string {
	return "HexaScoreBot/1.0 (+" + contactURL + ")"
}

type policy struct {
	hosts map[string]*host
	ua    string
	clock Clock
	// timeout bounds one attempt, from the moment it is sent until its body is closed.
	timeout time.Duration
}

// New builds the client.
func New(cfg Config) *Client {
	p := &policy{hosts: map[string]*host{}, ua: UserAgent(cfg.ContactURL), clock: cfg.Clock, timeout: cfg.Timeout}
	if p.clock == nil {
		p.clock = realClock{}
	}

	for name, l := range cfg.Hosts {
		p.hosts[name] = newHost(l)
	}

	if p.timeout <= 0 {
		p.timeout = timeout
	}

	// The clients get no overall Timeout: it would also cover the retry loop, so a Retry-After pause near 60 s
	// failed the request. Each attempt gets its own deadline in send instead.
	build := func(limit int64) *http.Client {
		if cfg.Base != nil {
			return &http.Client{Transport: p.wrap(cfg.Base)}
		}

		c := httpclient.Safe(httpclient.Options{
			Timeout: p.timeout, MaxBody: limit, Allow: cfg.Allow, WrapTransport: p.wrap,
		})
		c.Timeout = 0

		return c
	}

	return &Client{JSON: build(MaxJSONBytes), Image: build(MaxImageBytes), p: p}
}

func (p *policy) wrap(next http.RoundTripper) http.RoundTripper {
	return httpclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return p.roundTrip(next, req)
	})
}

func (p *policy) roundTrip(next http.RoundTripper, req *http.Request) (*http.Response, error) {
	h, ok := p.hosts[req.URL.Host]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrHostNotAllowed, req.URL.Host)
	}

	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", p.ua)

	ctx := req.Context()
	replayable := req.Body == nil || req.Body == http.NoBody || req.GetBody != nil

	for attempt := 1; ; attempt++ {
		resp, err := p.once(ctx, h, next, req)

		delay, pause, again := p.verdict(h, attempt, resp, err)
		if !again || attempt >= maxAttempts || !replayable {
			return resp, err
		}

		if resp != nil {
			_, _ = io.CopyN(io.Discard, resp.Body, 64<<10)
			_ = resp.Body.Close()
		}

		if pause > 0 {
			h.pause(p.clock, pause)
		} else if serr := p.clock.Sleep(ctx, delay); serr != nil {
			return nil, serr
		}

		if req.GetBody != nil {
			body, berr := req.GetBody()
			if berr != nil {
				return nil, berr
			}

			req = req.Clone(ctx)
			req.Body = body
		}
	}
}

func (p *policy) once(ctx context.Context, h *host, next http.RoundTripper, req *http.Request) (*http.Response, error) {
	h.mu.Lock()
	allowed, probe := h.br.allow(p.clock.Now())
	h.mu.Unlock()

	if !allowed {
		return nil, httpclient.ErrCircuitOpen
	}

	resp, err := p.send(ctx, h, next, req)

	h.mu.Lock()
	defer h.mu.Unlock()

	switch {
	case ctx.Err() != nil, errors.Is(err, httpclient.ErrBlockedAddress):
		if probe {
			h.br.probing = false
		}
	default:
		h.br.record(p.clock.Now(), err != nil || resp.StatusCode >= http.StatusInternalServerError, probe)
	}

	return resp, err
}

func (p *policy) send(ctx context.Context, h *host, next http.RoundTripper, req *http.Request) (*http.Response, error) {
	if err := h.wait(ctx, p.clock); err != nil {
		return nil, err
	}

	release, err := h.acquire(ctx)
	if err != nil {
		return nil, err
	}

	actx, cancel := context.WithTimeout(ctx, p.timeout)

	resp, err := next.RoundTrip(req.WithContext(actx))
	if err != nil {
		cancel()
		release()

		return nil, err
	}

	resp.Body = &releaseBody{ReadCloser: resp.Body, release: func() { release(); cancel() }}

	return resp, nil
}

// verdict decides whether the outcome is retried: after delay, or by
// pausing the whole host for pause when the server sent Retry-After.
func (p *policy) verdict(
	h *host,
	attempt int,
	resp *http.Response,
	err error,
) (delay, pause time.Duration, again bool) {
	backoff := min(time.Second<<attempt, maxBackoff)

	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, httpclient.ErrBlockedAddress) || errors.Is(err, httpclient.ErrCircuitOpen) ||
			errors.Is(err, ErrHostNotAllowed) {
			return 0, 0, false
		}

		return backoff, 0, true
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		h.rateLimited()
	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
	default:
		return 0, 0, false
	}

	if d, ok := parseRetryAfter(resp.Header.Get("Retry-After"), p.clock.Now()); ok {
		if d > maxRetryAfter {
			return 0, 0, false
		}

		return 0, max(d, time.Nanosecond), true
	}

	return backoff, 0, true
}

// parseRetryAfter reads delta-seconds or an HTTP date.
func parseRetryAfter(v string, now time.Time) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}

	if s, err := strconv.Atoi(v); err == nil && s >= 0 {
		return time.Duration(s) * time.Second, true
	}

	if t, err := http.ParseTime(v); err == nil {
		return max(t.Sub(now), 0), true
	}

	return 0, false
}
