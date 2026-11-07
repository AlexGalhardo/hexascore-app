package middleware

import "testing"

func TestNewsletterEndpoint(t *testing.T) {
	t.Parallel()

	cases := []struct{ base, slug, endpoint, origin string }{
		{
			NewsletterProdBase,
			NewsletterSlug,
			"https://email.hexascore.app/subscribe/hexascore",
			"https://email.hexascore.app",
		},
		{
			"http://localhost:3100/ignored/path",
			"hexascore",
			"http://localhost:3100/subscribe/hexascore",
			"http://localhost:3100",
		},
		{"", "hexascore", "", ""},
		{"javascript:alert(1)", "hexascore", "", ""},
		{"https://email.hexascore.app", "", "", ""},
	}

	for _, c := range cases {
		endpoint, origin := NewsletterEndpoint(c.base, c.slug)
		if endpoint != c.endpoint || origin != c.origin {
			t.Errorf("NewsletterEndpoint(%q, %q) = %q, %q", c.base, c.slug, endpoint, origin)
		}
	}
}
