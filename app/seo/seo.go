// Package seo builds the per-request SEO props shared with every page
// (SDD 09 §6.1, §7.7): canonical URL, hreflang alternates and robots.
package seo

import (
	"net/http"
	"strings"

	"github.com/AlexGalhardo/vukem/i18n"
)

// Locale pairs a locale code with its URL prefix ("" for the default).
type Locale struct {
	Code   string
	Prefix string
}

// OG holds the Open Graph values the client cannot derive.
type OG struct {
	Image string `json:"image,omitempty"`
	Type  string `json:"type"`
}

// Meta is the `seo` shared prop. Title and Description stay empty unless a
// controller overrides them; the client falls back to the app defaults.
type Meta struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Canonical   string            `json:"canonical"`
	Alternates  map[string]string `json:"alternates"`
	Robots      string            `json:"robots"`
	OG          OG                `json:"og"`
	JSONLD      []any             `json:"jsonLd,omitempty"`
}

// Config knows the public base URL and the locale prefixes. The first locale
// is the default: unprefixed, and the x-default alternate.
type Config struct {
	BaseURL string
	Locales []Locale
}

// privatePrefixes mirror the robots.txt Disallow list (SDD 09 §7.7).
var privatePrefixes = []string{
	"/admin", "/profile", "/me", "/billing", "/suggest", "/suggestions",
	"/library", "/lists", "/login", "/signup",
}

// Split returns the locale named by the first path segment and the path
// without that prefix; an unprefixed path belongs to the default locale.
func (c Config) Split(path string) (Locale, string) {
	first, rest, _ := strings.Cut(strings.TrimPrefix(path, "/"), "/")
	for _, l := range c.Locales[1:] {
		if strings.EqualFold(strings.Trim(l.Prefix, "/"), first) && first != "" {
			return l, "/" + rest
		}
	}

	if path == "" {
		path = "/"
	}

	return c.Locales[0], path
}

// For returns the default Meta for the request, whose path had the
// locale prefix stripped before routing (restored with i18n.LocalizePath).
func (c Config) For(r *http.Request) Meta {
	loc, path := c.Split(i18n.LocalizePath(r.Context(), r.URL.Path))
	base := strings.TrimRight(c.BaseURL, "/")

	alternates := make(map[string]string, len(c.Locales)+1)
	for _, l := range c.Locales {
		alternates[l.Code] = base + l.Prefix + path
	}

	alternates["x-default"] = alternates[c.Locales[0].Code]

	robots := "index,follow"
	for _, p := range privatePrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			robots = "noindex,nofollow"
		}
	}

	return Meta{
		Canonical:  alternates[loc.Code],
		Alternates: alternates,
		Robots:     robots,
		OG:         OG{Type: "website"},
	}
}

// Disallowed lists the robots.txt Disallow rules (SDD 09 §7.7).
func Disallowed() []string {
	rules := make([]string, 0, len(privatePrefixes)+1)
	rules = append(rules, privatePrefixes...)

	return append(rules, "/*?cursor=")
}
