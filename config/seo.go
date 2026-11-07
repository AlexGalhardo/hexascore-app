package config

// Indexable reports whether search engines may index this deployment: production only, unless
// SEO_NOINDEX=true. A deployment that runs APP_ENV=production but must stay out of search results (a preview, a local production build) sets SEO_NOINDEX.
func Indexable(production bool, get func(string) string) bool {
	return production && !NoindexForced(get)
}

// NoindexForced reports SEO_NOINDEX=true, which adds X-Robots-Tag: noindex, nofollow to every response.
func NoindexForced(get func(string) string) bool {
	return get("SEO_NOINDEX") == "true"
}
