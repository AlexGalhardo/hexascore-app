//go:build integration

package tests

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func ageUser(t *testing.T, app *vukemtest.App, id int64) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(),
		`UPDATE users SET created_at = now() - interval '30 days' WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
}

func suggestionStatus(t *testing.T, app *vukemtest.App, id string) string {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var s string
	if err := db.QueryRowContext(t.Context(), `SELECT status::text FROM suggestions WHERE id = $1`, id).
		Scan(&s); err != nil {
		t.Fatal(err)
	}

	return s
}

func TestCommunitySubmitWithdrawFlow(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	u := makeUser(t, app, "flow2@example.com", "user", true)
	ageUser(t, app, u.ID)
	item := communityItem(t, app, "public")
	c := app.Client(t).ActingAs("web", u)

	c.Get("/me/suggestions").AssertInertia("Community/MySuggestions").PropEquals("open", true)
	c.PostJSON("/suggestions", draftBody(item, "Mine", 0)).AssertStatus(http.StatusSeeOther)
	id := draftID(t, app, u.ID)

	c.Get("/suggestions/"+id).AssertInertia("Community/SuggestionStatus").
		PropEquals("suggestion.status", "draft").
		PropEquals("canAct", true).
		PropEquals("suggestion.events.0.to", "draft")
	c.Get("/me/suggestions").AssertInertia("Community/MySuggestions").
		PropEquals("suggestions.0.id", id).
		PropEquals("suggestions.0.title", "Mine")

	// A stale version is a 409 and changes nothing.
	c.PostJSON("/suggestions/"+id+"/submit", map[string]any{"version": 5}).AssertStatus(http.StatusConflict)
	if got := suggestionStatus(t, app, id); got != "draft" {
		t.Fatalf("status = %s after stale submit", got)
	}

	c.PostJSON("/suggestions/"+id+"/submit", map[string]any{"version": 1}).
		AssertStatus(http.StatusSeeOther).AssertRedirect("/suggestions/" + id)
	if got := suggestionStatus(t, app, id); got != "submitted" {
		t.Fatalf("status = %s, want submitted", got)
	}

	// A submitted suggestion cannot be submitted twice, but it can be withdrawn.
	c.PostJSON("/suggestions/"+id+"/submit", map[string]any{"version": 2}).AssertStatus(http.StatusConflict)
	c.PostJSON("/suggestions/"+id+"/withdraw", map[string]any{"version": 2}).AssertStatus(http.StatusSeeOther)
	c.Get("/suggestions/"+id).AssertInertia("Community/SuggestionStatus").
		PropEquals("suggestion.status", "withdrawn")
	c.PostJSON("/suggestions/"+id+"/withdraw", map[string]any{"version": 3}).AssertStatus(http.StatusConflict)
}

// TestSecurity_A01_CommunityStatusAuthorOrAdmin: the author and admins read a suggestion; another user, an anonymous
// caller and an unknown id all get 404, and nobody else can submit or withdraw it.
func TestSecurity_A01_CommunityStatusAuthorOrAdmin(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	author := makeUser(t, app, "sa@example.com", "user", true)
	other := makeUser(t, app, "so@example.com", "user", true)
	admin := makeUser(t, app, "sadm@example.com", models.RoleAdmin, true)
	ageUser(t, app, author.ID)
	ageUser(t, app, other.ID)
	item := communityItem(t, app, "public")

	app.Client(t).ActingAs("web", author).PostJSON("/suggestions", draftBody(item, "Private", 0)).
		AssertStatus(http.StatusSeeOther)
	id := draftID(t, app, author.ID)

	ac := app.Client(t).ActingAs("web", admin)
	ac.Get("/suggestions/"+id).AssertInertia("Community/SuggestionStatus").PropEquals("canAct", false)
	ac.PostJSON("/suggestions/"+id+"/submit", map[string]any{"version": 1}).AssertNotFound()
	ac.PostJSON("/suggestions/"+id+"/withdraw", map[string]any{"version": 1}).AssertNotFound()

	oc := app.Client(t).ActingAs("web", other)
	oc.Get("/suggestions/" + id).AssertNotFound()
	oc.Get("/suggestions/00000000-0000-7000-8000-000000000000").AssertNotFound()
	oc.Get("/suggestions/not-a-uuid").AssertNotFound()
	oc.PostJSON("/suggestions/"+id+"/submit", map[string]any{"version": 1}).AssertNotFound()
	oc.PostJSON("/suggestions/"+id+"/withdraw", map[string]any{"version": 1}).AssertNotFound()
	oc.Get("/me/suggestions").AssertInertia("Community/MySuggestions").PropEquals("suggestions", []any{})

	app.Client(t).Get("/suggestions/" + id).AssertRedirect("/login")
	app.Client(t).PostJSON("/suggestions/"+id+"/withdraw", map[string]any{"version": 1})

	if got := suggestionStatus(t, app, id); got != "draft" {
		t.Fatalf("status = %s after foreign attempts, want draft", got)
	}
}

// TestSecurity_A04_CommunitySubmitClosedByFlag: with the kill switch off, submit is 403
// community-submissions-closed and the draft stays a draft, while withdraw still works.
func TestSecurity_A04_CommunitySubmitClosedByFlag(t *testing.T) {
	t.Parallel()

	app := communityApp(t, false)
	u := makeUser(t, app, "closed@example.com", "user", true)
	ageUser(t, app, u.ID)
	item := communityItem(t, app, "public")

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	id := "0192b3c4-0000-7000-8000-00000000abcd"
	if _, err := db.ExecContext(t.Context(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes)
VALUES ($1, $2, 'edit', 'item', $3, 'movie', 'en', '{"title":{"after":"x"}}')`, id, u.ID, item); err != nil {
		t.Fatal(err)
	}

	c := app.Client(t).ActingAs("web", u)
	c.PostJSON("/suggestions/"+id+"/submit", map[string]any{"version": 1}).
		AssertForbidden().AssertSee("community-submissions-closed")
	if got := suggestionStatus(t, app, id); got != "draft" {
		t.Fatalf("status = %s, want draft", got)
	}

	c.PostJSON("/suggestions/"+id+"/withdraw", map[string]any{"version": 1}).AssertStatus(http.StatusSeeOther)
	if got := suggestionStatus(t, app, id); got != "withdrawn" {
		t.Fatalf("status = %s, want withdrawn", got)
	}
}
