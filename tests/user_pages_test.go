//go:build integration

package tests

import (
	"fmt"
	"testing"
)

// TestInertiaPages_UserLibraryLists runs under -run 'TestInertiaPages/(User|Library|Lists|Account)': one subtest
// per row of SDD 14 §6.2, each as owner, other user and anonymous where the row allows it.
func TestInertiaPages_UserLibraryLists(t *testing.T) {
	r := newListRig(t)
	owner, api := r.user("pgowner@example.com", "lists:write", "catalog:read")
	other, _ := r.user("pgother@example.com", "catalog:read")
	r.setNickname(owner, "PageOwner", false)
	r.setNickname(other, "PageOther", false)

	game := r.item("game", "public")
	r.status(owner.ID, game, "completed")

	create := func(title string, public bool) string {
		body := listBody("game", title)
		body["isPublic"] = public
		id, _ := api.PostJSON("/app/v1/lists", body).AssertStatus(201).JSONPath("id")

		return fmt.Sprint(id)
	}
	pub, priv := create("Shown", true), create("Hidden", false)
	api.PutJSON(fmt.Sprintf("/app/v1/lists/%s/items/%s", pub, game), map[string]any{}).AssertOK()

	own := browserFor(t, r.app, owner).Inertia("")
	oth := browserFor(t, r.app, other).Inertia("")
	anon := r.app.Client(t).Inertia("")

	t.Run("User/Profile", func(*testing.T) {
		own.Get("/u/pageowner").AssertOK().AssertInertia("User/Profile").
			PropEquals("profile.isOwner", true).HasProp("lists", 1)
		oth.Get("/u/pageowner").AssertOK().AssertInertia("User/Profile").PropEquals("profile.isOwner", false)
		anon.Get("/u/pageowner").AssertOK().AssertInertia("User/Profile").PropEquals("indexable", true)
	})

	t.Run("User/List", func(*testing.T) {
		own.Get("/u/pageowner/lists/"+pub).AssertOK().AssertInertia("User/List").
			PropEquals("canEdit", true).HasProp("items", 1)
		oth.Get("/u/pageowner/lists/"+pub).AssertOK().AssertInertia("User/List").PropEquals("canEdit", false).
			MissingProp("list.version").MissingProp("list.createdAt").MissingProp("items.0.addedAt").
			MissingProp("items.0.item.coverThumbKey")
		// One item is not indexable (SDD 14 FR-14-6 needs three); two more make it so.
		anon.Get("/u/pageowner/lists/"+pub).AssertOK().AssertInertia("User/List").
			PropEquals("canEdit", false).PropEquals("indexable", false)
		for range 2 {
			more := r.item("game", "public")
			api.PutJSON(fmt.Sprintf("/app/v1/lists/%s/items/%s", pub, more), map[string]any{}).AssertOK()
		}
		anon.Get("/u/pageowner/lists/"+pub).AssertOK().AssertInertia("User/List").
			PropEquals("canEdit", false).PropEquals("indexable", true)
		own.Get("/u/pageowner/lists/"+priv).AssertOK().AssertInertia("User/List").PropEquals("indexable", false)
		// A private list looks exactly like a missing one: 301 home (item 14), nothing else to tell them apart.
		oth.Get("/u/pageowner/lists/"+priv).AssertStatus(301).AssertHeader("Location", "/")
		anon.Get("/u/pageowner/lists/"+priv).AssertStatus(301).AssertHeader("Location", "/")
		anon.Get("/u/pageother/lists/"+pub).AssertStatus(301).AssertHeader("Location", "/")

		r.exec(`UPDATE users SET banned_at = now() WHERE id = $1`, owner.ID)
		anon.Get("/u/pageowner/lists/"+pub).AssertOK().AssertHeader("X-Robots-Tag", "noindex").
			AssertInertia("User/List").PropEquals("indexable", false)
		r.exec(`UPDATE users SET banned_at = NULL WHERE id = $1`, owner.ID)
	})

	t.Run("Library/Index", func(*testing.T) {
		own.Get("/library").AssertOK().AssertInertia("Library/Index").HasProp("groups").HasProp("counts")
		anon.Get("/library").AssertRedirect("/login")
	})

	t.Run("Lists/Index", func(*testing.T) {
		own.Get("/lists").AssertOK().AssertInertia("Lists/Index").HasProp("lists", 2).HasProp("limits")
		oth.Get("/lists").AssertOK().AssertInertia("Lists/Index").HasProp("lists", 0)
		anon.Get("/lists").AssertRedirect("/login")
	})

	t.Run("Lists/Edit", func(*testing.T) {
		own.Get("/lists/"+pub+"/edit").AssertOK().AssertInertia("Lists/Edit").
			HasProp("items", 3).PropEquals("version", 3)
		oth.Get("/lists/" + pub + "/edit").AssertNotFound()
		anon.Get("/lists/" + pub + "/edit").AssertRedirect("/login")
	})

	t.Run("Account/Notifications", func(t *testing.T) {
		sendNote(t, r, owner.ID, "pg-1")
		own.Get("/me/notifications").AssertOK().AssertInertia("Account/Notifications").HasProp("notifications.data", 1)
		oth.Get("/me/notifications").AssertOK().AssertInertia("Account/Notifications").
			HasProp("notifications.data", 0)
		anon.Get("/me/notifications").AssertRedirect("/login")
	})
}
