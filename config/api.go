package config

import "cmp"

// DefaultAPIHost serves the public API as https://api.hexascore.app/v1/... (SDD 05 §6.1).
const DefaultAPIHost = "api.hexascore.app"

// APIHost reads API_HOST through get (app.Source.Get).
func APIHost(get func(string) string) string {
	return cmp.Or(get("API_HOST"), DefaultAPIHost)
}
