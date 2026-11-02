package controllers

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/inertia"
)

func TestSearchShell(t *testing.T) {
	page := func(component string, props map[string]any) *inertia.Page {
		return &inertia.Page{Component: component, Props: props}
	}

	grouped := page("Search/Index", map[string]any{
		"results": []SearchHit{},
		"groups":  []SearchGroup{{Items: []SearchHit{{Title: `A <b> & "C"`, Path: "/games/a?x=1&y=2"}}}},
	})

	body := string(SearchShell(grouped).Body)
	for _, want := range []string{`<noscript>`, `href="/games/a?x=1&amp;y=2"`, `A &lt;b&gt; &amp; &#34;C&#34;`} {
		if !strings.Contains(body, want) {
			t.Errorf("body %q lacks %q", body, want)
		}
	}

	flat := page("Search/Index", map[string]any{"results": []SearchHit{{Title: "Zelda", Path: "/games/zelda"}}})
	if !strings.Contains(string(SearchShell(flat).Body), `href="/games/zelda"`) {
		t.Error("flat results not rendered")
	}

	for _, p := range []*inertia.Page{nil, page("Home/Index", map[string]any{"results": []SearchHit{{Path: "/x"}}}), page("Search/Index", map[string]any{})} {
		if got := SearchShell(p); got.Body != "" || got.Head != "" {
			t.Errorf("want empty shell, got %+v", got)
		}
	}
}
