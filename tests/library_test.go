//go:build integration

package tests

import (
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"

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
		PropEquals("seo.indexable", false).MissingProp("stats").MissingProp("ratings").MissingProp("lists")

	anon.Get("/u/openone").AssertOK().AssertHeaderMissing("X-Robots-Tag").AssertInertia("User/Profile").
		PropEquals("seo.indexable", true).HasProp("stats.statusByType.game.want")

	anon.Get("/u/nobody").AssertNotFound()

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
