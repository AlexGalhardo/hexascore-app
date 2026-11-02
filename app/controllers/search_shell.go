package controllers

import (
	"html"
	"html/template"
	"strings"

	"github.com/AlexGalhardo/vukem/inertia"
)

// SearchShell is inertia.Config.ClientShell. When the SSR render of Search/Index times out (the 3 s default
// on a cold Node process) the page falls back to a client shell; this keeps the result links in the first
// response for no-JS clients and link previews (W-05). Other components get an empty shell.
func SearchShell(p *inertia.Page) inertia.Shell {
	if p == nil || p.Component != "Search/Index" {
		return inertia.Shell{}
	}

	hits, _ := p.Props["results"].([]SearchHit)
	if len(hits) == 0 {
		groups, _ := p.Props["groups"].([]SearchGroup)
		for _, g := range groups {
			hits = append(hits, g.Items...)
		}
	}

	if len(hits) == 0 {
		return inertia.Shell{}
	}

	var b strings.Builder

	b.WriteString("<noscript><ul>")

	for _, h := range hits {
		b.WriteString(`<li><a href="` + html.EscapeString(h.Path) + `">` + html.EscapeString(h.Title) + "</a></li>")
	}

	b.WriteString("</ul></noscript>")

	return inertia.Shell{Body: template.HTML(b.String())} //nolint:gosec // every dynamic part is escaped above
}
