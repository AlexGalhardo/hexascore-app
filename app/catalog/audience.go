package catalog

import "slices"

// Audience selects which fields a read model may expose (SDD 01 §6.1).
type Audience string

const (
	// AudienceSite is the web app and the first-party app API: every field.
	AudienceSite Audience = "site"
	// AudienceAPI is the paid read API: only fields from allowed sources.
	AudienceAPI Audience = "api"
)

// Policy decides field visibility for one audience.
type Policy struct {
	audience Audience
	allowed  []string
}

// NewPolicy builds the policy; apiSources only matter for AudienceAPI.
func NewPolicy(audience Audience, apiSources []string) Policy {
	return Policy{audience: audience, allowed: apiSources}
}

// Allows reports whether a field from source may be shown. Fields with an
// unknown (empty) source are hidden from the API: fail closed (ADR 0006).
func (p Policy) Allows(source string) bool {
	if p.audience == AudienceSite {
		return true
	}
	return source != "" && slices.Contains(p.allowed, source)
}

// limitedSources may reach the API only for identifiers and titles until
// their licence note exists (open issue 01-2: Open Library, A-05).
var limitedSources = []string{"openlibrary"}

// AllowsContent is Allows for descriptive fields (synopsis, dates, counts,
// covers): limited sources are excluded for the API.
func (p Policy) AllowsContent(source string) bool {
	if p.audience == AudienceAPI && slices.Contains(limitedSources, source) {
		return false
	}
	return p.Allows(source)
}

// RelationSources is the item_entities sources a relation filter may match: nil (all) for the site, the
// allowlist for the API, so filtering or counting by a relation never reveals a restricted-source row
// (SDD 01 §6.1 "entity relations").
func (p Policy) RelationSources() []string {
	if p.audience == AudienceSite {
		return nil
	}

	return append([]string{}, p.allowed...)
}
