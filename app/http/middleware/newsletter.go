package middleware

import (
	"net/url"
	"strings"
)

const (
	// NewsletterProdBase is the vukem-email service that sends the HexaScore newsletter.
	NewsletterProdBase = "https://email.hexascore.app"
	// NewsletterDevBase is a local vukem-email run with PORT=3100 (3000 is this app).
	NewsletterDevBase = "http://localhost:3100"
	// NewsletterSlug is the vukem-email project slug of HexaScore.
	NewsletterSlug = "hexascore"
)

// NewsletterEndpoint returns the public subscribe URL the footer form posts to and the origin to allow in
// connect-src. The browser calls vukem-email directly: its per-IP limit (10 a minute) then counts visitors, not this
// server, and the email never passes through this app. An unusable base yields empty strings, which hides the form.
func NewsletterEndpoint(base, slug string) (endpoint, origin string) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || slug == "" {
		return "", ""
	}

	origin = u.Scheme + "://" + u.Host

	return origin + "/subscribe/" + url.PathEscape(slug), origin
}
