package controllers

import (
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
)

// legalUpdatedAt is the effective date of the legal texts in resources/lang.
// Bump it together with any change to the policy or terms copy.
const legalUpdatedAt = "2026-11-01"

// Legal renders /terms and /policy; the text lives in the catalogues under
// the slug's namespace (owner reviews the copy, Q-D-007).
func Legal(slug string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		return inertia.Render(c, "Legal/Page", inertia.Props{"slug": slug, "updatedAt": legalUpdatedAt})
	}
}
