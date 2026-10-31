//go:build integration

package tests

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

var communityQID atomic.Int32

func communityApp(t *testing.T, open bool) *vukemtest.App {
	t.Helper()

	env := map[string]string{"DATABASE_URL": pgtest.Database(t, "")}
	if open {
		env["FLAG_COMMUNITY_SUBMISSIONS"] = "true"
	}

	return vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
}

func communityItem(t *testing.T, app *vukemtest.App, visibility string) string {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	id := uuid.NewV7().String()
	if _, err := db.ExecContext(
		t.Context(),
		`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, 'movie', $2, $3, $4)`,
		id, "s-"+id, fmt.Sprintf("Q%d", 7000+communityQID.Add(1)), visibility,
	); err != nil {
		t.Fatal(err)
	}

	return id
}

func draftBody(itemID, title string, version int) map[string]any {
	return map[string]any{
		"targetKind": "item",
		"targetId":   itemID,
		"version":    version,
		"locale":     "en",
		"sourceUrls": []string{"https://www.wikidata.org/wiki/Q127367"},
		"changes":    map[string]any{"title": map[string]any{"before": nil, "after": title}},
	}
}

func TestCommunityFlagOff(t *testing.T) {
	t.Parallel()

	app := communityApp(t, false)
	u := makeUser(t, app, "flag-off@example.com", "user", true)
	item := communityItem(t, app, "public")
	c := app.Client(t).ActingAs("web", u)

	c.Get("/suggest/item/"+item).AssertInertia("Community/SuggestEdit").PropEquals("open", false)
	c.PostJSON("/suggestions", draftBody(item, "X", 0)).
		AssertForbidden().
		AssertSee("community-submissions-closed")
}

func TestCommunityDraftFlow(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	u := makeUser(t, app, "flow@example.com", "user", true)
	item := communityItem(t, app, "public")
	c := app.Client(t).ActingAs("web", u)

	c.Get("/suggest/item/"+item).AssertInertia("Community/SuggestEdit").
		PropEquals("open", true).
		PropEquals("target.domain", "movie").
		PropEquals("quota.left", float64(3)).
		PropEquals("draft", nil)

	c.PostJSON("/suggestions", draftBody(item, "First", 0)).AssertStatus(http.StatusSeeOther)

	c.Get("/suggest/item/"+item).AssertInertia("Community/SuggestEdit").
		PropEquals("draft.version", float64(1)).
		PropEquals("draft.changes.title.after", "First")

	id := draftID(t, app, u.ID)
	c.PatchJSON("/suggestions/"+id, draftBody(item, "Second", 1)).AssertStatus(http.StatusSeeOther)
	c.Get("/suggest/item/"+item).AssertInertia("Community/SuggestEdit").
		PropEquals("draft.version", float64(2)).
		PropEquals("draft.changes.title.after", "Second")

	c.PatchJSON("/suggestions/"+id, draftBody(item, "Third", 1)).AssertStatus(http.StatusConflict)

	bad := draftBody(item, "Fourth", 2)
	bad["locale"] = "xx"
	c.PatchJSON("/suggestions/"+id, bad).AssertUnprocessable()
}

func draftID(t *testing.T, app *vukemtest.App, authorID int64) string {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var id string
	if err := db.QueryRowContext(t.Context(), `SELECT id FROM suggestions WHERE author_id = $1`, authorID).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

// TestSecurity_A01_CommunityDraftAuthorOnly: another user's draft, an unknown or non-public target and an
// anonymous caller never reach the draft (SDD 04 §6).
func TestSecurity_A01_CommunityDraftAuthorOnly(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	owner := makeUser(t, app, "owner@example.com", "user", true)
	other := makeUser(t, app, "other@example.com", "user", true)
	unverified := makeUser(t, app, "unverified@example.com", "user", false)
	item := communityItem(t, app, "public")
	hidden := communityItem(t, app, "hidden")

	oc := app.Client(t).ActingAs("web", owner)
	oc.PostJSON("/suggestions", draftBody(item, "Mine", 0)).AssertStatus(http.StatusSeeOther)
	id := draftID(t, app, owner.ID)

	app.Client(t).ActingAs("web", other).PatchJSON("/suggestions/"+id, draftBody(item, "Stolen", 1)).AssertNotFound()
	app.Client(t).ActingAs("web", other).Get("/suggest/item/"+item).AssertInertia("Community/SuggestEdit").
		PropEquals("draft", nil)

	oc.PostJSON("/suggestions", draftBody(uuid.NewV7().String(), "Ghost", 0)).AssertNotFound()
	oc.PostJSON("/suggestions", draftBody(hidden, "Hidden", 0)).AssertNotFound()
	oc.Get("/suggest/item/" + hidden).AssertNotFound()

	app.Client(t).PostJSON("/suggestions", draftBody(item, "Anon", 0))
	if n := draftCount(t, app); n != 1 {
		t.Fatalf("%d drafts, want 1", n)
	}

	app.Client(t).ActingAs("web", unverified).PostJSON("/suggestions", draftBody(item, "Unverified", 0))
	if n := draftCount(t, app); n != 1 {
		t.Fatalf("unverified save created a draft, %d drafts", n)
	}
}

func draftCount(t *testing.T, app *vukemtest.App) int {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM suggestions`).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

// TestSecurity_A04_CommunityHoneypot: a filled website field is a generic 422 and stores nothing.
func TestSecurity_A04_CommunityHoneypot(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	u := makeUser(t, app, "bot@example.com", "user", true)
	item := communityItem(t, app, "public")

	body := draftBody(item, "Spam", 0)
	body["website"] = "https://spam.example"
	app.Client(t).ActingAs("web", u).PostJSON("/suggestions", body).AssertUnprocessable()

	if n := draftCount(t, app); n != 0 {
		t.Fatalf("honeypot hit stored %d drafts", n)
	}
}
