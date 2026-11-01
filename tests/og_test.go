//go:build integration

package tests

import (
	"bytes"
	"image/png"
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/seo"
)

func (r *listRig) publicList(owner models.User, itemID string, public bool) string {
	r.t.Helper()

	id := uuid.NewV7().String()
	r.exec(`INSERT INTO lists (id, user_id, type, title, is_public, item_count) VALUES ($1, $2, 'game', 'Best', $3, 1)`,
		id, owner.ID, public)
	r.exec(`INSERT INTO list_items (list_id, item_id, position) VALUES ($1, $2, 1)`, id, itemID)

	return id
}

func TestOGImage(t *testing.T) {
	r := newSocialRig(t, true)
	pub, pc := r.member("og-pub@example.com", "Sharer", false)
	priv, _ := r.member("og-priv@example.com", "Secret", true)

	g := r.releasedItem("game")
	r.rate(pc, g, 8)
	hidden := r.item("game", "hidden")
	list := r.publicList(pub, g, true)
	privateList := r.publicList(pub, g, false)
	ownerPrivateList := r.publicList(priv, g, true)

	c := r.app.Client(t)
	for _, path := range []string{"/og/items/" + g + ".png", "/og/profiles/sharer.png", "/og/lists/" + list + ".png"} {
		res := c.Get(path).AssertOK().
			AssertHeader("Content-Type", "image/png").
			AssertHeader("Cache-Control", "public, max-age=86400")

		img, err := png.Decode(bytes.NewReader(res.Recorder().Body.Bytes()))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if b := img.Bounds(); b.Dx() != seo.CardWidth || b.Dy() != seo.CardHeight {
			t.Errorf("%s: size %v", path, b)
		}
	}

	// The ROADMAP acceptance uses curl -sI (HEAD).
	c.Do(http.MethodHead, "/og/items/"+g+".png", nil, "").AssertOK().AssertHeader("Content-Type", "image/png")

	for _, path := range []string{
		"/og/items/" + hidden + ".png",
		"/og/items/not-a-uuid.png",
		"/og/items/" + g,
		"/og/profiles/secret.png",
		"/og/profiles/nobody.png",
		"/og/lists/" + privateList + ".png",
		"/og/lists/" + ownerPrivateList + ".png",
		"/og/users/" + g + ".png",
	} {
		c.Get(path).AssertNotFound()
	}

	page := c.Inertia("").Get(r.itemPath(g)).AssertOK().AssertInertia("Catalog/Show")
	if v, _ := page.Prop("ogImage"); v != "/og/items/"+g+".png" {
		t.Errorf("item ogImage = %v", v)
	}
	page = c.Inertia("").Get("/u/sharer").AssertOK().AssertInertia("User/Profile")
	if v, _ := page.Prop("ogImage"); v != "/og/profiles/sharer.png" {
		t.Errorf("profile ogImage = %v", v)
	}
	page = c.Inertia("").Get("/u/sharer/lists/" + list).AssertOK().AssertInertia("User/List")
	if v, _ := page.Prop("ogImage"); v != "/og/lists/"+list+".png" {
		t.Errorf("list ogImage = %v", v)
	}
	page = r.app.Client(t).Inertia("").Get("/u/secret").AssertOK().AssertInertia("User/Profile")
	if v, ok := page.Prop("ogImage"); ok {
		t.Errorf("private profile ogImage = %v, want absent", v)
	}
}

func TestOGImage_FlagOff(t *testing.T) {
	r := newSocialRig(t, false)
	g := r.releasedItem("game")
	c := r.app.Client(t)

	c.Get("/og/items/" + g + ".png").AssertNotFound()

	page := c.Inertia("").Get(r.itemPath(g)).AssertOK().AssertInertia("Catalog/Show")
	if v, ok := page.Prop("ogImage"); ok {
		t.Errorf("ogImage = %v with the flag off", v)
	}
}
