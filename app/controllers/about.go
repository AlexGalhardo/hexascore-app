package controllers

import (
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
)

// About renders the /about page; its copy lives in the "about." catalogue namespace.
func About(c *web.Ctx) error {
	return inertia.Render(c, "About/Index", inertia.Props{})
}
