package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/http/problems"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestProblemCatalog: every code of SDD 05 §6.2 has a type URL that redirects to /docs/errors#<code> and a
// section on the errors page.
func TestProblemCatalog(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	doc, err := os.ReadFile("../resources/docs/errors.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range problems.Catalog {
		app.Client(t).Get("/problems/" + p.Code).AssertStatus(302).AssertRedirect("/docs/errors#" + p.Code)

		if !strings.Contains(string(doc), "\n### "+p.Code+"\n") {
			t.Errorf("resources/docs/errors.md has no section for %q", p.Code)
		}
	}

	app.Client(t).Get("/problems/nope").AssertNotFound()
}
