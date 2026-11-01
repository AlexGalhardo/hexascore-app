// Package botverify tells a real search crawler from a client that only borrows its User-Agent: reverse
// DNS of the client IP must end in a domain the operator documents, and the forward lookup of that name
// must return the same IP (ADR 0022, SDD 13 §7.1.1).
package botverify

import (
	"cmp"
	"context"
	"errors"
	"net"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
)

// Crawler is a bot the site allows (Q-A18). Domains are the reverse DNS suffixes its operator documents;
// an empty list means the operator publishes no rDNS method, so the bot is matched by UA only and keeps
// the anonymous page limits.
type Crawler struct {
	Name    string
	Tokens  []string // lowercase User-Agent substrings
	Domains []string
}

// Crawlers is the allow list, search engines first. Google-Extended and Applebot-Extended are robots.txt
// tokens only: Googlebot and Applebot do that crawling.
func Crawlers() []Crawler {
	return []Crawler{
		{
			"Googlebot",
			[]string{
				"googlebot",
				"google-inspectiontool",
				"googleother",
				"adsbot-google",
				"storebot-google",
				"mediapartners-google",
			},
			// Google lists googleusercontent.com for user-triggered fetchers, but bc.googleusercontent.com is any
			// Google Cloud customer VM, so only the gae. subdomain Google's fetchers use is accepted.
			[]string{"googlebot.com", "google.com", "gae.googleusercontent.com"},
		},
		{"Bingbot", []string{"bingbot", "adidxbot", "bingpreview", "microsoftpreview"}, []string{"search.msn.com"}},
		{"Applebot", []string{"applebot"}, []string{"applebot.apple.com"}},
		{
			"YandexBot",
			[]string{"yandexbot", "yandeximages", "yandexmobilebot"},
			[]string{"yandex.ru", "yandex.net", "yandex.com"},
		},
		// shortcut: the operators below publish IP lists, not rDNS; they are UA-matched and rate-limited
		// like anonymous clients. Fetch their JSON lists on a schedule if they need the exemption.
		{"DuckDuckBot", []string{"duckduckbot", "duckassistbot"}, nil},
		{"GPTBot", []string{"gptbot", "oai-searchbot", "chatgpt-user"}, nil},
		{"ClaudeBot", []string{"claudebot", "claude-user", "claude-searchbot"}, nil},
		{"CCBot", []string{"ccbot"}, nil},
		{"PerplexityBot", []string{"perplexitybot", "perplexity-user"}, nil},
	}
}

// Claimed returns the allowed crawler whose token appears in ua.
func Claimed(ua string) (Crawler, bool) {
	ua = strings.ToLower(ua)

	for _, c := range Crawlers() {
		for _, tok := range c.Tokens {
			if strings.Contains(ua, tok) {
				return c, true
			}
		}
	}

	return Crawler{}, false
}

// Verdict is the outcome of a check.
type Verdict uint8

const (
	// Unverified means no rDNS method for this crawler, or DNS did not answer in time. The client is treated
	// as an anonymous visitor.
	Unverified Verdict = iota
	// Verified means reverse and forward DNS agree on a documented domain.
	Verified
	// Spoofed means DNS answered and the IP does not belong to the crawler it names.
	Spoofed
)

// Resolver is the part of *net.Resolver the check needs; tests pass a fake.
type Resolver interface {
	LookupAddr(ctx context.Context, addr string) ([]string, error)
	LookupHost(ctx context.Context, host string) ([]string, error)
}

const (
	// DefaultTimeout bounds both lookups together, so a slow DNS server never holds a page request longer.
	DefaultTimeout = 1500 * time.Millisecond
	settledTTL     = 24 * time.Hour
	// A DNS failure is retried soon, but not on every request of a burst.
	unverifiedTTL = time.Minute
)

// Verifier checks claims and caches the verdict per crawler and IP in a shared store.
type Verifier struct {
	Resolver Resolver
	Cache    *cache.Cache
	Timeout  time.Duration
}

// Verify reports whether ip belongs to crawler c.
func (v *Verifier) Verify(ctx context.Context, c Crawler, ip string) Verdict {
	if len(c.Domains) == 0 {
		return Unverified
	}

	key := "botverify:" + c.Name + ":" + ip
	// The cache only saves lookups: a store error falls through to a fresh check.
	if got, ok, err := cache.Get[Verdict](ctx, v.Cache, key); err == nil && ok {
		return got
	}

	// shortcut: concurrent first requests from one IP each resolve; add a singleflight if crawlers burst.
	verdict := v.lookup(ctx, c, ip)

	ttl := settledTTL
	if verdict == Unverified {
		ttl = unverifiedTTL
	}

	_ = v.Cache.Put(ctx, key, verdict, ttl)

	return verdict
}

func (v *Verifier) lookup(ctx context.Context, c Crawler, ip string) Verdict {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return Spoofed
	}

	ctx, cancel := context.WithTimeout(ctx, cmp.Or(v.Timeout, DefaultTimeout))
	defer cancel()

	names, err := v.Resolver.LookupAddr(ctx, addr.String())
	if err != nil {
		return failed(err)
	}

	for _, name := range names {
		host := strings.TrimSuffix(strings.ToLower(name), ".")
		if !inDomains(host, c.Domains) {
			continue
		}

		addrs, err := v.Resolver.LookupHost(ctx, host)
		if err != nil {
			return failed(err)
		}

		if slices.ContainsFunc(addrs, func(a string) bool {
			p, err := netip.ParseAddr(a)
			return err == nil && p.Unmap() == addr.Unmap()
		}) {
			return Verified
		}
	}

	return Spoofed
}

// failed maps a lookup error: "no such name" is a definite answer, anything else (timeout, SERVFAIL) is not.
func failed(err error) Verdict {
	if dnsErr, ok := errors.AsType[*net.DNSError](err); ok && dnsErr.IsNotFound {
		return Spoofed
	}

	return Unverified
}

func inDomains(host string, domains []string) bool {
	return slices.ContainsFunc(domains, func(d string) bool { return strings.HasSuffix(host, "."+d) })
}
