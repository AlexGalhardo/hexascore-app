package offers_test

import (
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/offers"
)

func TestPlatformURLFindsTheServiceSite(t *testing.T) {
	for name, want := range map[string]string{
		"Netflix":                   "https://www.netflix.com",
		"Netflix Standard with Ads": "https://www.netflix.com",
		"Amazon Prime Video":        "https://www.primevideo.com",
		"Amazon Video":              "https://www.primevideo.com",
		"Telecine Amazon Channel":   "https://www.primevideo.com",
		"Apple TV Store":            "https://tv.apple.com",
		"Google Play Movies":        "https://play.google.com/store/movies",
		"YouTube":                   "https://www.youtube.com/movies",
		"HBO Max":                   "https://www.max.com",
		"Max":                       "https://www.max.com",
		"Peacock Premium Plus":      "https://www.peacocktv.com",
		"Claro video":               "https://www.clarovideo.com",
		"Movistar Plus+":            "https://www.movistarplus.es",
		"Fandango At Home":          "https://www.fandangoathome.com",
		"Rakuten TV":                "https://rakuten.tv",
		"Universal+ Amazon Channel": "https://www.primevideo.com",
	} {
		if got, ok := offers.PlatformURL(name); !ok || got != want {
			t.Errorf("%q = %q (%v), want %q", name, got, ok, want)
		}
	}

	for _, name := range []string{"Some Local Cinema Club", "Maximum Movies", ""} {
		if got, ok := offers.PlatformURL(name); ok {
			t.Errorf("%q has no known site but resolved to %q", name, got)
		}
	}
}
