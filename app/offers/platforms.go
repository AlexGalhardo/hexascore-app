package offers

import (
	"regexp"
	"strings"
)

// platform maps a streaming service name to the service's own site. TMDB's watch providers give a name and an id but
// no link to the service, only a link to its own watch page, so a link to the service is looked up here, and a service
// not listed has no link: it is left out of the page rather than sent to TMDB (owner 2026-11-04). The first match wins.
type platform struct {
	re  *regexp.Regexp
	url string
}

func words(pattern string) *regexp.Regexp {
	return regexp.MustCompile(`(^|[^a-z0-9])(` + pattern + `)([^a-z0-9]|$)`)
}

var platforms = []platform{
	// Channels sold inside Prime Video ("Telecine Amazon Channel") are bought and played there.
	{words(`amazon[ a-z]*channel|amazon|prime video`), "https://www.primevideo.com"},
	{words(`netflix`), "https://www.netflix.com"},
	{words(`apple tv`), "https://tv.apple.com"},
	{words(`google play`), "https://play.google.com/store/movies"},
	{words(`youtube`), "https://www.youtube.com/movies"},
	{words(`disney`), "https://www.disneyplus.com"},
	{words(`hbo max|max`), "https://www.max.com"},
	{words(`paramount`), "https://www.paramountplus.com"},
	{words(`peacock`), "https://www.peacocktv.com"},
	{words(`hulu`), "https://www.hulu.com"},
	{words(`globoplay`), "https://globoplay.globo.com"},
	{words(`claro video`), "https://www.clarovideo.com"},
	{words(`telecine`), "https://www.telecine.com.br"},
	{words(`rakuten`), "https://rakuten.tv"},
	{words(`movistar`), "https://www.movistarplus.es"},
	{words(`skyshowtime`), "https://www.skyshowtime.com"},
	{words(`fandango`), "https://www.fandangoathome.com"},
	{words(`mubi`), "https://mubi.com"},
	{words(`crunchyroll`), "https://www.crunchyroll.com"},
	{words(`tubi`), "https://tubitv.com"},
	{words(`pluto tv`), "https://pluto.tv"},
	{words(`plex`), "https://watch.plex.tv"},
	{words(`starz`), "https://www.starz.com"},
	{words(`universal\+?`), "https://www.universalplus.com"},
}

// PlatformURL is the site of a streaming service by its TMDB provider name, or false when the service is not known.
func PlatformURL(name string) (string, bool) {
	n := strings.ToLower(strings.TrimSpace(name))

	for _, p := range platforms {
		if p.re.MatchString(n) {
			return p.url, true
		}
	}

	return "", false
}
