package config

// Indexable reports whether search engines may index this deployment: production only, unless
// SEO_NOINDEX=true. Staging runs APP_ENV=production (SDD 11 §6.4), so it sets SEO_NOINDEX.
func Indexable(production bool, get func(string) string) bool {
	return production && get("SEO_NOINDEX") != "true"
}
