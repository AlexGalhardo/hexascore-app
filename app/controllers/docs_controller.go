package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
)

type docPage struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Body  string `json:"body"`
	TOC   []struct {
		ID    string `json:"id"`
		Text  string `json:"text"`
		Level int    `json:"level"`
	} `json:"toc"`
}

// Docs serves /docs and /docs/{slug} from the pages scripts/build-docs.ts
// rendered and escaped at build time; unknown slugs fall through to the 404 page.
func Docs(raw []byte) (web.HandlerFunc, error) {
	var pages []docPage
	if err := json.Unmarshal(raw, &pages); err != nil {
		return nil, fmt.Errorf("docs: %w", err)
	}

	bySlug := make(map[string]docPage, len(pages))
	slugs := make([]string, 0, len(pages))

	for _, p := range pages {
		bySlug[p.Slug] = p
		slugs = append(slugs, p.Slug)
	}

	return func(c *web.Ctx) error {
		slug := c.Param("slug")
		if slug == "" {
			slug = "index"
		}

		p, ok := bySlug[slug]
		if !ok || (slug == "index" && c.Param("slug") != "") {
			return NotFound(c)
		}

		return inertia.Render(c, "Docs/Page", inertia.Props{
			"slug": p.Slug, "title": p.Title, "body": p.Body, "toc": p.TOC, "pages": slugs,
		})
	}, nil
}
