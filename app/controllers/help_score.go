package controllers

import (
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// HelpScore renders the public "how scores work" page (SDD 02 §10); the
// category keys come from scoring so the page cannot drift from the formula.
func HelpScore(c *web.Ctx) error {
	return inertia.Render(c, "Help/Score", inertia.Props{"categoryKeys": scoring.CategoryKeys})
}
