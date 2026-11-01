//go:build integration

package tests

import (
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func (r *listRig) exec(q string, args ...any) {
	r.t.Helper()

	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}

	if _, err := db.ExecContext(r.t.Context(), q, args...); err != nil {
		r.t.Fatal(err)
	}
}

func (r *listRig) edition(workID string) string {
	r.t.Helper()

	r.n++
	r.exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x') ON CONFLICT DO NOTHING`,
	)
	id := uuid.NewV7().String()
	r.exec(`INSERT INTO editions (id, work_id, openlibrary_id, title, language, source)
VALUES ($1, $2, $3, 'Ed', 'en', 'hexascore')`, id, workID, fmt.Sprintf("OL%dM", r.n))

	return id
}

func (r *listRig) status(userID int64, itemID, status string) {
	r.t.Helper()
	r.exec(`INSERT INTO item_status (user_id, item_id, status) VALUES ($1, $2, $3)`, userID, itemID, status)
}

func (r *listRig) setNickname(u models.User, nick string, private bool) {
	r.t.Helper()
	r.exec(`UPDATE users SET username = lower($2), display_username = $2, profile_private = $3 WHERE id = $1`,
		u.ID, nick, private)
}

func TestLibraryGroups(t *testing.T) {
	r := newListRig(t)
	u, api := r.user("lib@example.com", "catalog:read", "status:write")
	other, _ := r.user("libother@example.com", "catalog:read")

	// 31 games to cross the 30 per type page, one movie, one hidden game and a foreign row.
	for range 31 {
		r.status(u.ID, r.item("game", "public"), "completed")
	}

	r.status(u.ID, r.item("movie", "public"), "want")
	r.status(u.ID, r.item("game", "hidden"), "want")
	r.status(other.ID, r.item("book", "public"), "want")

	var view struct {
		Counts map[string]map[string]int `json:"counts"`
		Groups []struct {
			Type  string           `json:"type"`
			Total int              `json:"total"`
			Items []map[string]any `json:"items"`
			Next  int              `json:"next"`
		} `json:"groups"`
	}
	api.GetJSON("/app/v1/me/library").AssertOK().JSON(&view)

	if len(view.Groups) != 2 || view.Groups[0].Type != "game" || view.Groups[1].Type != "movie" {
		t.Fatalf("groups = %+v; want game then movie", view.Groups)
	}

	if g := view.Groups[0]; g.Total != 31 || len(g.Items) != 30 || g.Next != 30 {
		t.Fatalf("game group total=%d items=%d next=%d; want 31/30/30", g.Total, len(g.Items), g.Next)
	}

	if view.Counts["completed"]["game"] != 31 || view.Counts["want"]["movie"] != 1 {
		t.Fatalf("counts = %v", view.Counts)
	}

	api.GetJSON("/app/v1/me/library?type=game&cursor=30").AssertOK().JSON(&view)
	if len(view.Groups) != 1 || len(view.Groups[0].Items) != 1 || view.Groups[0].Next != -1 {
		t.Fatalf("second page = %+v", view.Groups)
	}

	api.GetJSON("/app/v1/me/library?status=bogus").AssertUnprocessable()
	api.GetJSON("/app/v1/me/library?status=want").AssertOK()

	web := browserFor(t, r.app, u)
	web.Get("/library").AssertOK()
}

func TestEditionPickerList(t *testing.T) {
	r := newListRig(t)
	u, _ := r.user("picker@example.com", "catalog:read", "status:write")
	web := browserFor(t, r.app, u)

	book, game := r.item("book", "public"), r.item("game", "public")
	ed := r.edition(book)
	r.edition(r.item("book", "public"))

	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	web.GetJSON("/items/" + book + "/editions").AssertOK().JSON(&body)
	if len(body.Data) != 1 || body.Data[0].ID != ed {
		t.Fatalf("editions = %+v; want only %s", body.Data, ed)
	}

	var page struct {
		Data []struct{} `json:"data"`
		Next int        `json:"next"`
	}
	web.GetJSON("/items/" + book + "/editions").AssertOK().JSON(&page)
	if page.Next != -1 {
		t.Fatalf("next = %d; want -1 on a single page", page.Next)
	}

	many := r.item("book", "public")
	for range 11 {
		r.edition(many)
	}
	web.GetJSON("/items/" + many + "/editions").AssertOK().JSON(&page)
	if len(page.Data) != 10 || page.Next != 10 {
		t.Fatalf("page 1 = %d next %d; want 10 next 10", len(page.Data), page.Next)
	}
	web.GetJSON("/items/" + many + "/editions?cursor=10").AssertOK().JSON(&page)
	if len(page.Data) != 1 || page.Next != -1 {
		t.Fatalf("page 2 = %d next %d; want 1 next -1", len(page.Data), page.Next)
	}

	web.GetJSON("/items/" + game + "/editions").AssertNotFound()
	web.GetJSON("/items/" + r.item("book", "hidden") + "/editions").AssertNotFound()
	web.GetJSON("/items/not-a-uuid/editions").AssertNotFound()
	r.app.Client(t).GetJSON("/items/" + book + "/editions").AssertUnauthorized()
}

func TestEditionNotOfWork(t *testing.T) {
	r := newListRig(t)
	u, api := r.user("ed@example.com", "catalog:read", "status:write")
	_, stranger := r.user("edother@example.com", "catalog:read")

	book, otherBook := r.item("book", "public"), r.item("book", "public")
	ed, foreign := r.edition(book), r.edition(otherBook)
	r.status(u.ID, book, "completed")

	api.PutJSON("/app/v1/items/"+book+"/edition", map[string]any{"editionId": foreign}).AssertUnprocessable()
	api.PutJSON("/app/v1/items/"+book+"/edition", map[string]any{"editionId": uuid.NewV7().String()}).
		AssertUnprocessable()
	api.PutJSON("/app/v1/items/"+uuid.NewV7().String()+"/edition", map[string]any{"editionId": ed}).AssertNotFound()
	stranger.PutJSON("/app/v1/items/"+book+"/edition", map[string]any{"editionId": ed}).AssertForbidden()

	api.PutJSON("/app/v1/items/"+book+"/edition", map[string]any{"editionId": ed}).AssertOK()
	api.PutJSON("/app/v1/items/"+book+"/edition", map[string]any{"editionId": ed}).AssertOK()

	var view struct {
		Groups []struct {
			Items []struct {
				Edition struct {
					ID string `json:"id"`
				} `json:"edition"`
			} `json:"items"`
		} `json:"groups"`
	}
	api.GetJSON("/app/v1/me/library").AssertOK().JSON(&view)

	if got := view.Groups[0].Items[0].Edition.ID; got != ed {
		t.Fatalf("edition = %q; want %q", got, ed)
	}

	api.Delete("/app/v1/items/" + book + "/edition").AssertNoContent()
	api.Delete("/app/v1/items/" + book + "/edition").AssertNoContent()
}

func TestPublicProfilePrivacy(t *testing.T) {
	r := newListRig(t)
	priv, _ := r.user("priv@example.com", "catalog:read")
	pub, _ := r.user("pub@example.com", "catalog:read")
	_, viewer := r.user("viewer@example.com", "catalog:read")
	r.setNickname(priv, "SecretOne", true)
	r.setNickname(pub, "OpenOne", false)
	r.status(priv.ID, r.item("game", "public"), "want")
	r.status(pub.ID, r.item("game", "public"), "want")

	anon := r.app.Client(t).Inertia("")
	anon.Get("/u/secretone").AssertOK().AssertHeader("X-Robots-Tag", "noindex").AssertInertia("User/Profile").
		PropEquals("profile.nickname", "SecretOne").PropEquals("profile.isPrivate", true).
		PropEquals("indexable", false).MissingProp("stats").MissingProp("ratings").MissingProp("lists")

	anon.Get("/u/openone").AssertOK().AssertHeaderMissing("X-Robots-Tag").AssertInertia("User/Profile").
		PropEquals("indexable", true).HasProp("stats.statusByType.game.want")

	anon.Get("/u/nobody").AssertStatus(301).AssertHeader("Location", "/")

	r.exec(`UPDATE users SET banned_at = now() WHERE id = $1`, pub.ID)
	anon.Get("/u/openone").AssertOK().AssertHeader("X-Robots-Tag", "noindex").AssertInertia("User/Profile").
		PropEquals("indexable", false).HasProp("stats")
	r.exec(`UPDATE users SET banned_at = NULL WHERE id = $1`, pub.ID)

	var body struct {
		Profile map[string]any `json:"profile"`
		Stats   any            `json:"stats"`
	}
	viewer.GetJSON("/app/v1/users/SecretOne").AssertOK().JSON(&body)
	if body.Stats != nil || body.Profile["nickname"] != "SecretOne" {
		t.Fatalf("private API profile = %+v", body)
	}

	viewer.GetJSON("/app/v1/users/openone").AssertOK().JSON(&body)
	if body.Stats == nil {
		t.Fatal("public API profile has no stats")
	}

	owner := browserFor(t, r.app, priv).Inertia("")
	owner.Get("/u/secretone").AssertOK().AssertInertia("User/Profile").HasProp("stats")
}

// TestMergeMovesUserEditions: a chosen edition follows its work into the merge target; where the user already
// picked an edition of the target, the target's choice wins (SDD 14 §8).
func TestMergeMovesUserEditions(t *testing.T) {
	r := newListRig(t)
	a, _ := r.user("eda@example.com")
	b, _ := r.user("edb@example.com")
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	q, err := r.app.Queue()
	if err != nil {
		t.Fatal(err)
	}

	from, into := r.item("book", "public"), r.item("book", "public")
	edFrom, edInto := r.edition(from), r.edition(into)
	r.exec(`INSERT INTO user_item_editions (user_id, item_id, edition_id)
VALUES ($1, $2, $3), ($4, $2, $3), ($4, $5, $6)`, a.ID, from, edFrom, b.ID, into, edInto)

	if err := catalog.Merge(t.Context(), db, q, from, into, 0); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		user    int64
		edition string
	}{{a.ID, edFrom}, {b.ID, edInto}} {
		if got := queryInts(t, r, `SELECT count(*) FROM user_item_editions WHERE user_id = $1 AND item_id = $2
AND edition_id = $3`, c.user, into, c.edition); got[0] != 1 {
			t.Errorf("user %d: edition %s on target missing", c.user, c.edition)
		}
	}
	if got := queryInts(t, r, `SELECT count(*) FROM user_item_editions WHERE item_id = $1 OR user_id = $2 AND
edition_id <> $3`, from, b.ID, edInto); got[0] != 0 {
		t.Errorf("%d stale edition rows", got[0])
	}
}
