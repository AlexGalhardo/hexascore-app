//go:build integration

package tests

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"
)

func createBody(title string, version int) map[string]any {
	return map[string]any{
		"kind":       "create",
		"domain":     "movie",
		"version":    version,
		"locale":     "en",
		"sourceUrls": []string{"https://www.wikidata.org/wiki/Q127367"},
		"changes":    map[string]any{"title": map[string]any{"before": nil, "after": title}},
	}
}

func mergeBody(src, dst string, version int) map[string]any {
	return map[string]any{
		"kind":       "merge",
		"targetKind": "item",
		"targetId":   src,
		"version":    version,
		"locale":     "en",
		"changes":    map[string]any{"*merge*": map[string]any{"before": src, "after": dst}},
	}
}

func TestCommunityCreateFlow(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	u := makeUser(t, app, "create@example.com", "user", true)
	c := app.Client(t).ActingAs("web", u)

	c.Get("/suggest/new/movie").AssertInertia("Community/SuggestCreate").
		PropEquals("open", true).
		PropEquals("domain", "movie").
		PropEquals("kind", "create").
		PropEquals("quota.left", float64(3)).
		PropEquals("draft", nil)
	c.Get("/suggest/new/podcast").AssertNotFound()

	c.PostJSON("/suggestions", createBody("Brand new", 0)).AssertRedirect("/suggest/new/movie?kind=create")
	c.Get("/suggest/new/movie").AssertInertia("Community/SuggestCreate").
		PropEquals("draft.version", float64(1)).
		PropEquals("draft.changes.title.after", "Brand new")

	id := draftID(t, app, u.ID)
	c.PatchJSON("/suggestions/"+id, createBody("Renamed", 1)).AssertStatus(http.StatusSeeOther)
	c.Get("/suggest/new/movie").AssertInertia("Community/SuggestCreate").
		PropEquals("draft.changes.title.after", "Renamed")

	missing := createBody("", 0)
	missing["changes"] = map[string]any{}
	c.PostJSON("/suggestions", missing).AssertUnprocessable()

	bad := createBody("X", 0)
	bad["domain"] = "podcast"
	c.PostJSON("/suggestions", bad).AssertUnprocessable()

	withTarget := createBody("Y", 0)
	withTarget["targetId"] = communityItem(t, app, "public")
	c.PostJSON("/suggestions", withTarget).AssertStatus(http.StatusSeeOther)
	db, _ := app.DB()
	var n int
	if err := db.QueryRowContext(
		t.Context(),
		`SELECT count(*) FROM suggestions WHERE kind = 'create' AND target_id IS NOT NULL`,
	).Scan(&n); err != nil || n != 0 {
		t.Fatalf("create draft with target_id: n=%d err=%v", n, err)
	}
}

func TestCommunityMergeRequest(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	u := makeUser(t, app, "merge@example.com", "user", true)
	c := app.Client(t).ActingAs("web", u)
	src, dst := communityItem(t, app, "public"), communityItem(t, app, "public")

	c.Get("/suggest/new/movie?kind=merge").AssertInertia("Community/SuggestCreate").PropEquals("kind", "merge")
	c.PostJSON("/suggestions", mergeBody(src, dst, 0)).AssertStatus(http.StatusSeeOther)

	// Same item, unknown item, hidden item and a different type are all refused without storing anything.
	hidden := communityItem(t, app, "hidden")
	db, _ := app.DB()
	book := uuid.NewV7().String()
	if _, err := db.ExecContext(
		t.Context(),
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'book', $2, 'Q9990001')`,
		book,
		"b-"+book,
	); err != nil {
		t.Fatal(err)
	}
	before := draftCount(t, app)
	c.PostJSON("/suggestions", mergeBody(src, src, 0)).AssertUnprocessable()
	c.PostJSON("/suggestions", mergeBody(src, uuid.NewV7().String(), 0)).AssertUnprocessable()
	c.PostJSON("/suggestions", mergeBody(src, hidden, 0)).AssertUnprocessable()
	c.PostJSON("/suggestions", mergeBody(src, book, 0)).AssertUnprocessable()
	c.PostJSON("/suggestions", mergeBody(src, "not-a-uuid", 0)).AssertUnprocessable()
	if n := draftCount(t, app); n != before {
		t.Fatalf("refused merge stored drafts: %d -> %d", before, n)
	}
}

// TestSecurity_A01_CommunityCreateClosedAndAuth: the create page and create/merge saves are closed by the flag and
// need a verified account.
func TestSecurity_A01_CommunityCreateClosedAndAuth(t *testing.T) {
	t.Parallel()

	off := communityApp(t, false)
	u := makeUser(t, off, "off@example.com", "user", true)
	oc := off.Client(t).ActingAs("web", u)
	oc.Get("/suggest/new/movie").AssertInertia("Community/SuggestCreate").PropEquals("open", false)
	oc.PostJSON("/suggestions", createBody("X", 0)).AssertForbidden().AssertSee("community-submissions-closed")

	on := communityApp(t, true)
	unverified := makeUser(t, on, "unv@example.com", "user", false)
	on.Client(t).Get("/suggest/new/movie").AssertRedirect("/login")
	on.Client(t).ActingAs("web", unverified).PostJSON("/suggestions", createBody("X", 0))
	on.Client(t).PostJSON("/suggestions", createBody("X", 0))
	if n := draftCount(t, on); n != 0 {
		t.Fatalf("unverified or anonymous create stored %d drafts", n)
	}
}

// TestSecurity_A01_CommunityKindFixedAtCreate: a draft's kind cannot be switched by an update (a create draft cannot
// be patched into a merge or edit), and another user's create draft is a 404.
func TestSecurity_A01_CommunityKindFixedAtCreate(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	a := makeUser(t, app, "ka@example.com", "user", true)
	b := makeUser(t, app, "kb@example.com", "user", true)
	ac := app.Client(t).ActingAs("web", a)
	ac.PostJSON("/suggestions", createBody("Mine", 0)).AssertStatus(http.StatusSeeOther)
	id := draftID(t, app, a.ID)

	src, dst := communityItem(t, app, "public"), communityItem(t, app, "public")
	ac.PatchJSON("/suggestions/"+id, mergeBody(src, dst, 1)).AssertUnprocessable()
	app.Client(t).ActingAs("web", b).PatchJSON("/suggestions/"+id, createBody("Stolen", 1)).AssertNotFound()
}
