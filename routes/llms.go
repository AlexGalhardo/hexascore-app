package routes

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/llmstxt"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/resources"
)

// LLMs serves /llms.txt and /llms-full.txt (https://vukem.dev/docs/digging-deeper/llms-txt): what HexaScore is, the
// public catalogue sections, the paid API docs and the legal pages. llms-full.txt streams the API docs Markdown
// only, never item pages.
func LLMs(r *router.Router, app *vukem.App) error {
	raw, err := resources.DocsJSON()
	if err != nil {
		return err
	}

	doc, err := llmsDoc(strings.TrimSuffix(cmp.Or(app.Source.Get("APP_URL"), "https://hexascore.app"), "/"), raw)
	if err != nil {
		return err
	}

	return llmstxt.Register(r, doc)
}

func llmsDoc(site string, docsJSON []byte) (llmstxt.Doc, error) {
	var pages []struct{ Slug, Title string }
	if err := json.Unmarshal(docsJSON, &pages); err != nil {
		return llmstxt.Doc{}, fmt.Errorf("llms.txt: docs.json: %w", err)
	}

	docs := []llmstxt.Link{
		{Title: "API reference", URL: site + "/docs/reference", Notes: "every endpoint (OpenAPI 3.1)"},
	}
	for _, p := range pages {
		docs = append(docs, llmstxt.Link{
			Title: p.Title,
			URL:   site + "/docs/" + p.Slug,
			Content: func(context.Context) (string, error) {
				b, err := resources.DocsMarkdown(p.Slug)

				return string(b), err
			},
		})
	}

	section := func(path, title, notes string) llmstxt.Link {
		return llmstxt.Link{Title: title, URL: site + path, Notes: notes}
	}

	return llmstxt.Doc{
		Title:   "HexaScore",
		Summary: "Community scores for games, movies, TV series and books, with a paid read-only JSON API.",
		Details: "Members rate each item from 0 to 10 in six categories of its type; the score of an item is the " +
			"community mean. People, companies, franchises, platforms, genres, game modes, countries and languages " +
			"have a page under each category, such as /games/platforms/playstation-5 or /movies/people/ridley-scott. " +
			"The site is also served in Brazilian Portuguese (/pt-br/) and Spanish (/es/).",
		Sections: []llmstxt.Section{
			{Title: "Catalogue", Links: []llmstxt.Link{
				section("/games", "Games", "community scores and categories of video games"),
				section("/movies", "Movies", "community scores of films"),
				section("/series", "Series", "community scores of TV series"),
				section("/books", "Books", "community scores of books and their editions"),
				section("/how-score-works", "How scores work", "the public scoring formula and its rules"),
			}},
			{Title: "API", Links: docs},
			{Title: "Plans and policies", Links: []llmstxt.Link{
				section("/about", "About", "what HexaScore is, how it scores and where its code lives"),
				section("/pricing", "Pricing", "API plans and quotas"),
				section("/terms", "Terms of use", ""),
				section("/privacy", "Privacy policy", ""),
			}},
		},
	}, nil
}
