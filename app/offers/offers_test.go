package offers_test

import (
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/offers"
)

func TestDefaultCountry(t *testing.T) {
	t.Parallel()

	for locale, want := range map[string]string{"pt-BR": "BR", "es": "ES", "en": "US", "": "US", "fr": "US"} {
		if got := offers.DefaultCountry(locale); got != want {
			t.Errorf("DefaultCountry(%q) = %q, want %q", locale, got, want)
		}
	}
}

func TestEnvTagger(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"AFFILIATE_AMAZON_BR_TAG": "hexascore-20",
		"AFFILIATE_AMAZON_US_TAG": "bad tag&x=1", // would add a parameter
		"AFFILIATE_AMAZON_TAG":    "",
		"AFFILIATE_GOG_TAG":       "a/b",
	}
	tag := offers.EnvTagger(func(k string) string { return env[k] })

	if got := tag("amazon_br"); got != "hexascore-20" {
		t.Errorf("amazon_br = %q", got)
	}

	for _, store := range []string{"amazon_us", "amazon", "gog", "steam"} {
		if got := tag(store); got != "" {
			t.Errorf("%s = %q, want no tag (unset or unsafe)", store, got)
		}
	}
}

func TestWithTag(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, url, param, tag, want string }{
		{
			"appended", "https://www.amazon.com.br/s?k=9780261103573", "tag", "hexascore-20",
			"https://www.amazon.com.br/s?k=9780261103573&tag=hexascore-20",
		},
		{"no tag configured", "https://www.amazon.com.br/s?k=1", "tag", "", "https://www.amazon.com.br/s?k=1"},
		{"store without a programme", "https://store.steampowered.com/app/1", "", "x", "https://store.steampowered.com/app/1"},
		{"replaces an existing tag", "https://www.amazon.com/dp/B0?tag=old-20", "tag", "new-20", "https://www.amazon.com/dp/B0?tag=new-20"},
		{"no query yet", "https://www.amazon.com/dp/B0", "tag", "new-20", "https://www.amazon.com/dp/B0?tag=new-20"},
		{"unparseable stays as it is", "://nope", "tag", "x", "://nope"},
	} {
		if got := offers.WithTag(tc.url, tc.param, tc.tag); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}
