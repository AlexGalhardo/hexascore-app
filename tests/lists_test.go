//go:build integration

package tests

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type listRig struct {
	t   *testing.T
	app *vukemtest.App
	acc *accounts.Accounts
	n   int
}

func newListRig(t *testing.T) *listRig {
	t.Helper()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	return &listRig{t: t, app: app, acc: acc}
}

// user makes a verified user and a token client with the given abilities.
func (r *listRig) user(email string, abilities ...string) (models.User, *vukemtest.Client) {
	r.t.Helper()

	u := makeUser(r.t, r.app, email, "user", true)
	plain, _, err := r.acc.Tokens.Create(r.t.Context(), u.AuthID(), "t", abilities, time.Hour)
	if err != nil {
		r.t.Fatal(err)
	}

	return u, r.app.Client(r.t).WithToken(plain)
}

func (r *listRig) item(typ, visibility string) string {
	r.t.Helper()

	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}

	r.n++
	id := uuid.NewV7().String()
	if _, err := db.ExecContext(
		r.t.Context(),
		`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, $2, $3, $4, $5)`,
		id,
		typ,
		"s-"+id,
		fmt.Sprintf("Q%d", 9000+r.n),
		visibility,
	); err != nil {
		r.t.Fatal(err)
	}

	return id
}

func listBody(typ, title string) map[string]any {
	return map[string]any{"type": typ, "title": title, "isPublic": true}
}

func TestSecurity_A01_ListOwnerOnly(t *testing.T) {
	r := newListRig(t)
	_, owner := r.user("owner@example.com", "lists:write")
	stranger, other := r.user("stranger@example.com", "lists:write")
	_, noAbility := r.user("readonly@example.com", "catalog:read")

	res := owner.PostJSON("/app/v1/lists", listBody("game", "Mine")).AssertStatus(201)
	id, _ := res.JSONPath("id")
	path := fmt.Sprintf("/app/v1/lists/%v", id)

	other.PatchJSON(path, map[string]any{"title": "Hijack"}).AssertNotFound()
	other.Delete(path).AssertNotFound()
	noAbility.PatchJSON(path, map[string]any{"title": "x"}).AssertForbidden()
	r.app.Client(t).PatchJSON(path, map[string]any{"title": "x"}).AssertUnauthorized()
	// An unknown list answers exactly like someone else's.
	other.PatchJSON("/app/v1/lists/"+uuid.NewV7().String(), map[string]any{"title": "x"}).AssertNotFound()
	other.PatchJSON("/app/v1/lists/not-a-uuid", map[string]any{"title": "x"}).AssertNotFound()

	vukemtest.AssertDatabaseHas(t, r.app, "lists", map[string]any{"id": id, "title": "Mine"})

	// The same through the session guard.
	browserFor(
		t,
		r.app,
		stranger,
	).WithHeader("Referer", "/lists").
		PostJSON("/lists", listBody("game", "web")).
		AssertRedirect("/lists")
	browserFor(
		t,
		r.app,
		stranger,
	).WithHeader("Referer", "/lists").
		PatchJSON(fmt.Sprintf("/lists/%v", id), map[string]any{"title": "Hijack"}).
		AssertNotFound()
	vukemtest.AssertDatabaseHas(t, r.app, "lists", map[string]any{"id": id, "title": "Mine"})

	owner.PatchJSON(path, map[string]any{"title": "Renamed", "isPublic": false, "description": "d"}).AssertOK().
		AssertJSONPath("title", "Renamed")
	vukemtest.AssertDatabaseHas(t, r.app, "lists", map[string]any{"id": id, "title": "Renamed", "is_public": false})
	owner.PatchJSON(path, map[string]any{"description": ""}).AssertOK()
	vukemtest.AssertDatabaseHas(t, r.app, "lists", map[string]any{"id": id, "description": nil})

	owner.Delete(path).AssertNoContent()
	vukemtest.AssertDatabaseMissing(t, r.app, "lists", map[string]any{"id": id})
}

func TestListLimits(t *testing.T) {
	r := newListRig(t)
	_, c := r.user("limits@example.com", "lists:write")
	_, other := r.user("limits2@example.com", "lists:write")

	for i := range library.MaxListsPerType {
		c.PostJSON("/app/v1/lists", listBody("game", fmt.Sprintf("L%d", i))).AssertStatus(201)
	}
	c.PostJSON("/app/v1/lists", listBody("game", "one too many")).AssertStatus(409)
	c.PostJSON("/app/v1/lists", listBody("movie", "other type")).AssertStatus(201)
	other.PostJSON("/app/v1/lists", listBody("game", "other user")).AssertStatus(201)

	for name, body := range map[string]map[string]any{
		"empty title":  listBody("book", ""),
		"long title":   listBody("book", strings.Repeat("x", library.MaxTitle+1)),
		"long desc":    {"type": "book", "title": "t", "description": strings.Repeat("x", library.MaxDescription+1)},
		"unknown type": listBody("music", "t"),
		"bad first id": {"type": "book", "title": "t", "firstItemId": "nope"},
	} {
		t.Run(name, func(*testing.T) { c.PostJSON("/app/v1/lists", body).AssertUnprocessable() })
	}
	c.PostJSON("/app/v1/lists", listBody("book", strings.Repeat("x", library.MaxTitle))).AssertStatus(201)
}

func TestListTypeMismatch(t *testing.T) {
	r := newListRig(t)
	_, c := r.user("mismatch@example.com", "lists:write")
	game, movie := r.item("game", "public"), r.item("movie", "public")

	first := func(item string) map[string]any {
		b := listBody("game", "With item")
		b["firstItemId"] = item

		return b
	}
	c.PostJSON("/app/v1/lists", first(movie)).AssertUnprocessable()
	c.PostJSON("/app/v1/lists", first(uuid.NewV7().String())).AssertNotFound()
	c.PostJSON("/app/v1/lists", first(r.item("game", "takedown"))).AssertNotFound()
	vukemtest.AssertDatabaseMissing(t, r.app, "lists", map[string]any{"title": "With item"})

	res := c.PostJSON("/app/v1/lists", first(game)).AssertStatus(201)
	res.AssertJSONPath("itemCount", float64(1))
	vukemtest.AssertDatabaseHas(t, r.app, "list_items", map[string]any{"item_id": game, "position": 1})
}

func TestListWriteIdempotencyKey(t *testing.T) {
	r := newListRig(t)
	_, c := r.user("idem@example.com", "lists:write")

	c = c.WithHeader("Idempotency-Key", "k-1")
	first := c.PostJSON("/app/v1/lists", listBody("game", "Once")).AssertStatus(201)
	id, _ := first.JSONPath("id")
	again := c.PostJSON("/app/v1/lists", listBody("game", "Once")).AssertStatus(201)
	again.AssertJSONPath("id", id)

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM lists WHERE title = 'Once'`).
		Scan(&n); err != nil ||
		n != 1 {
		t.Fatalf("lists titled Once: %d (err %v), want 1", n, err)
	}
}

func TestListTypeIsImmutableAndEventsFire(t *testing.T) {
	r := newListRig(t)
	u, _ := r.user("events@example.com", "lists:write")
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	bus := events.New(slog.Default(), 0)
	events.Listen(bus, func(_ context.Context, e library.ListVisibilityChanged) error {
		got = append(got, fmt.Sprintf("visibility:%t", e.Visible))
		return nil
	})
	events.Listen(
		bus,
		func(_ context.Context, _ library.ListDeleted) error { got = append(got, "deleted"); return nil },
	)
	events.Listen(
		bus,
		func(_ context.Context, _ library.ListItemAdded) error { got = append(got, "added"); return nil },
	)

	svc := &library.Lists{DB: db, Events: bus, Now: time.Now}
	l, err := svc.Create(
		t.Context(),
		u.ID,
		library.CreateInput{Type: "game", Title: "E", IsPublic: true, FirstItemID: r.item("game", "public")},
	)
	if err != nil {
		t.Fatal(err)
	}
	private := false
	if l, err = svc.Update(t.Context(), l, library.UpdateInput{IsPublic: &private}); err != nil {
		t.Fatal(err)
	}
	title := "Same visibility"
	if l, err = svc.Update(t.Context(), l, library.UpdateInput{Title: &title}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `UPDATE lists SET type = 'movie' WHERE id = $1`, l.ID); err == nil {
		t.Error("lists.type changed")
	}
	if err := svc.Delete(t.Context(), l); err != nil {
		t.Fatal(err)
	}

	if want := "added,visibility:false,deleted"; strings.Join(got, ",") != want {
		t.Errorf("events %v, want %s", got, want)
	}
}
