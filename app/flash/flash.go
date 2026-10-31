// Package flash queues the "status" toast the layout shows after a redirect (SDD 09 §7.9).
package flash

import (
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
)

// Status flashes the resources/lang message of key in the request locale. Without a translation bundle (unit
// tests) it flashes the key itself.
func Status(c *web.Ctx, key string) {
	msg := key
	if b := i18n.BundleFrom(c.Context()); b != nil {
		msg = b.T(i18n.Locale(c.Context()), key)
	}

	inertia.Flash(c, "status", msg)
}
