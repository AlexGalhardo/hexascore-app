package controllers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AlexGalhardo/vukem/i18n"
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

type docSet struct {
	bySlug map[string]docPage
	slugs  []string
}

func parseDocSet(raw []byte) (docSet, error) {
	var pages []docPage
	if err := json.Unmarshal(raw, &pages); err != nil {
		return docSet{}, err
	}

	set := docSet{bySlug: make(map[string]docPage, len(pages)), slugs: make([]string, 0, len(pages))}

	for _, p := range pages {
		set.bySlug[p.Slug] = p
		set.slugs = append(set.slugs, p.Slug)
	}

	return set, nil
}

// Docs serves /docs and /docs/{slug} from the pages scripts/build-docs.ts
// rendered and escaped at build time; unknown slugs fall through to the 404 page.
// raws maps a locale to its docs.json; "en" is required and is the fallback for
// a locale or a page without a translation. Slugs and the sidebar order are the
// English ones in every locale. The coverage page also gets the measured
// percentages (P3.B.3.2).
func Docs(raws map[string][]byte, coverage APICoverage) (web.HandlerFunc, error) {
	sets := make(map[string]docSet, len(raws))

	for locale, raw := range raws {
		set, err := parseDocSet(raw)
		if err != nil {
			return nil, fmt.Errorf("docs %s: %w", locale, err)
		}

		sets[locale] = set
	}

	en, ok := sets["en"]
	if !ok {
		return nil, errors.New("docs: missing en pages")
	}

	return func(c *web.Ctx) error {
		slug := c.Param("slug")
		if slug == "" {
			slug = "index"
		}

		p, ok := en.bySlug[slug]
		if !ok || (slug == "index" && c.Param("slug") != "") {
			return NotFound(c)
		}

		if tr, found := sets[i18n.Locale(c.Context())].bySlug[slug]; found {
			p = tr
		}

		props := inertia.Props{"slug": p.Slug, "title": p.Title, "body": p.Body, "toc": p.TOC, "pages": en.slugs}
		if slug == "coverage" {
			props["coverage"] = coverage
		}

		return inertia.Render(c, "Docs/Page", props)
	}, nil
}
