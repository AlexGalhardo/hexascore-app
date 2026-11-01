package botverify_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/cache"

	"github.com/AlexGalhardo/hexascore-app/app/botverify"
)

// fakeResolver answers from maps; no live DNS in tests.
type fakeResolver struct {
	ptr     map[string][]string
	forward map[string][]string
	calls   int
	block   bool
}

func (f *fakeResolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	f.calls++

	if f.block {
		<-ctx.Done()

		return nil, ctx.Err()
	}

	names, ok := f.ptr[addr]
	if !ok {
		return nil, &net.DNSError{Err: "no such host", Name: addr, IsNotFound: true}
	}

	return names, nil
}

func (f *fakeResolver) LookupHost(_ context.Context, host string) ([]string, error) {
	addrs, ok := f.forward[host]
	if !ok {
		return nil, &net.DNSError{Err: "no such host", Name: host, IsNotFound: true}
	}

	return addrs, nil
}

func verifier(r botverify.Resolver) *botverify.Verifier {
	return &botverify.Verifier{Resolver: r, Cache: cache.New(cache.NewMemory(100), ""), Timeout: 50 * time.Millisecond}
}

const googlebotUA = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

func TestClaimed(t *testing.T) {
	t.Parallel()

	for ua, want := range map[string]string{
		googlebotUA: "Googlebot",
		"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)":                                     "Bingbot",
		"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)":      "GPTBot",
		"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; ClaudeBot/1.0; +claudebot@anthropic.com)":     "ClaudeBot",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0 Safari/537.36": "",
	} {
		c, ok := botverify.Claimed(ua)
		if c.Name != want || ok != (want != "") {
			t.Errorf("Claimed(%q) = %q, %v; want %q", ua, c.Name, ok, want)
		}
	}
}

func TestVerify(t *testing.T) {
	t.Parallel()

	google, _ := botverify.Claimed(googlebotUA)
	gpt, _ := botverify.Claimed("GPTBot/1.2")

	r := &fakeResolver{
		ptr: map[string][]string{
			"66.249.66.1":  {"crawl-66-249-66-1.googlebot.com."},
			"203.0.113.9":  {"evil.example.net."},
			"203.0.113.10": {"crawl-1.googlebot.com.evil.net."},
			// A Google Cloud customer VM: googleusercontent.com, but not a Google crawler.
			"34.1.2.3":     {"3.2.1.34.bc.googleusercontent.com."},
			"203.0.113.11": {"crawl-66-249-66-2.googlebot.com."},
		},
		forward: map[string][]string{
			"crawl-66-249-66-1.googlebot.com":   {"66.249.66.1"},
			"3.2.1.34.bc.googleusercontent.com": {"34.1.2.3"},
			// The PTR is forged: the forward name points elsewhere.
			"crawl-66-249-66-2.googlebot.com": {"66.249.66.2"},
		},
	}
	v := verifier(r)
	ctx := t.Context()

	for ip, want := range map[string]botverify.Verdict{
		"66.249.66.1":  botverify.Verified,
		"203.0.113.9":  botverify.Spoofed,
		"203.0.113.10": botverify.Spoofed,
		"34.1.2.3":     botverify.Spoofed,
		"203.0.113.11": botverify.Spoofed,
		"198.51.100.1": botverify.Spoofed, // no PTR at all
	} {
		if got := v.Verify(ctx, google, ip); got != want {
			t.Errorf("Verify(Googlebot, %s) = %v, want %v", ip, got, want)
		}
	}

	// Operators that publish no rDNS method are never verified, only UA-matched.
	if got := v.Verify(ctx, gpt, "66.249.66.1"); got != botverify.Unverified {
		t.Errorf("GPTBot = %v, want Unverified", got)
	}

	// Verdicts are cached: a second check does not resolve again.
	calls := r.calls
	v.Verify(ctx, google, "66.249.66.1")

	if r.calls != calls {
		t.Errorf("cached verdict resolved again (%d -> %d calls)", calls, r.calls)
	}
}

// A resolver that hangs is cut by the timeout and the client is unverified, not spoofed.
func TestVerifyTimeout(t *testing.T) {
	t.Parallel()

	google, _ := botverify.Claimed(googlebotUA)
	v := verifier(&fakeResolver{block: true})

	start := time.Now()
	if got := v.Verify(t.Context(), google, "66.249.66.1"); got != botverify.Unverified {
		t.Errorf("hung resolver = %v, want Unverified", got)
	}

	if d := time.Since(start); d > time.Second {
		t.Errorf("lookup took %v, want the 50ms timeout", d)
	}
}

func TestVerifyStoreDown(t *testing.T) {
	t.Parallel()

	google, _ := botverify.Claimed(googlebotUA)
	v := &botverify.Verifier{
		Resolver: &fakeResolver{
			ptr:     map[string][]string{"66.249.66.1": {"crawl.googlebot.com."}},
			forward: map[string][]string{"crawl.googlebot.com": {"66.249.66.1"}},
		},
		Cache:   cache.New(brokenStore{}, ""),
		Timeout: time.Second,
	}

	// The cache is an optimisation: a broken store still gets a fresh verdict.
	if got := v.Verify(t.Context(), google, "66.249.66.1"); got != botverify.Verified {
		t.Errorf("store down = %v, want Verified", got)
	}
}

type brokenStore struct{ cache.Store }

var errDown = errors.New("store down")

func (brokenStore) Get(context.Context, string) ([]byte, bool, error) { return nil, false, errDown }

func (brokenStore) Set(context.Context, string, []byte, time.Duration) error { return errDown }
