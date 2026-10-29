//go:build integration

package tests

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/library"
)

func createList(t *testing.T, c *vukemtest.Client, typ string) (id string, path string) {
	t.Helper()

	v, _ := c.PostJSON("/app/v1/lists", listBody(typ, "L")).AssertStatus(201).JSONPath("id")
	id = fmt.Sprint(v)

	return id, "/app/v1/lists/" + id
}

func queryInts(t *testing.T, r *listRig, q string, args ...any) []int {
	t.Helper()

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.QueryContext(t.Context(), q, args...)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	var out []int
	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		out = append(out, n)
	}

	return out
}

func TestListItemsAddRemoveReorder(t *testing.T) {
	r := newListRig(t)
	_, c := r.user("items@example.com", "lists:write")
	_, other := r.user("items2@example.com", "lists:write")
	a, b, d := r.item("game", "public"), r.item("game", "public"), r.item("game", "public")
	movie := r.item("movie", "public")
	id, path := createList(t, c, "game")

	c.PutJSON(path+"/items/"+a, map[string]any{"note": "first"}).AssertOK().AssertJSONPath("itemCount", float64(1))
	c.PutJSON(path+"/items/"+b, map[string]any{}).AssertOK()
	c.PutJSON(path+"/items/"+d, map[string]any{}).AssertOK().AssertJSONPath("version", float64(3))
	// Re-adding is a no-op 200: no new row, no version bump.
	c.PutJSON(path+"/items/"+a, map[string]any{"note": "ignored"}).AssertOK().AssertJSONPath("version", float64(3))
	vukemtest.AssertDatabaseHas(t, r.app, "list_items", map[string]any{"item_id": a, "note": "first", "position": 1})

	c.PutJSON(path+"/items/"+movie, map[string]any{}).AssertUnprocessable()
	c.PutJSON(path+"/items/"+r.item("game", "hidden"), map[string]any{}).AssertNotFound()
	c.PutJSON(path+"/items/"+a, map[string]any{"note": strings.Repeat("x", library.MaxNote+1)}).AssertUnprocessable()
	other.PutJSON(path+"/items/"+r.item("game", "public"), map[string]any{}).AssertNotFound()
	other.PutJSON(path+"/order", map[string]any{"version": 3, "itemIds": []string{a, b, d}}).AssertNotFound()

	// Reorder: stale version 409, changed set 422, then the new order is stored densely.
	c.PutJSON(path+"/order", map[string]any{"version": 2, "itemIds": []string{d, b, a}}).AssertStatus(409)
	c.PutJSON(path+"/order", map[string]any{"version": 3, "itemIds": []string{d, b}}).AssertUnprocessable()
	c.PutJSON(path+"/order", map[string]any{"version": 3, "itemIds": []string{d, b, b}}).AssertUnprocessable()
	c.PutJSON(path+"/order", map[string]any{"version": 3, "itemIds": []string{d, b, a}}).
		AssertOK().
		AssertJSONPath("version", float64(4))
	vukemtest.AssertDatabaseHas(t, r.app, "list_items", map[string]any{"item_id": d, "position": 1})
	vukemtest.AssertDatabaseHas(t, r.app, "list_items", map[string]any{"item_id": a, "position": 3})
	// The previous version is now stale.
	c.PutJSON(path+"/order", map[string]any{"version": 3, "itemIds": []string{a, b, d}}).AssertStatus(409)

	c.Delete(path + "/items/" + b).AssertNoContent()
	c.Delete(path + "/items/" + b).AssertNoContent()
	if got := queryInts(t, r, `SELECT item_count FROM lists WHERE id = $1`, id); len(got) != 1 || got[0] != 2 {
		t.Errorf("item_count after remove = %v, want [2]", got)
	}
	if got := queryInts(
		t,
		r,
		`SELECT position FROM list_items WHERE list_id = $1 ORDER BY position`,
		id,
	); fmt.Sprint(
		got,
	) != "[1 3]" {
		t.Errorf("positions after remove = %v, want a gap [1 3]", got)
	}
}

func TestListItemLimit(t *testing.T) {
	r := newListRig(t)
	u, c := r.user("full@example.com", "lists:write")
	id, path := createList(t, c, "game")
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO items (id, type, slug, wikidata_qid)
SELECT gen_random_uuid(), 'game', 'bulk-' || g, 'Q' || (700000 + g) FROM generate_series(1, $1) g`, library.MaxItemsPerList); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `INSERT INTO list_items (list_id, item_id, position)
SELECT $1, id, row_number() OVER () FROM items WHERE slug LIKE 'bulk-%'`, id); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE lists SET item_count = $2 WHERE id = $1 AND user_id = $3`,
		id,
		library.MaxItemsPerList,
		u.ID,
	); err != nil {
		t.Fatal(err)
	}

	c.PutJSON(path+"/items/"+r.item("game", "public"), map[string]any{}).AssertStatus(409)
}

// TestListEventsAfterCommit: list.item_added exists only when the transaction commits.
func TestListEventsAfterCommit(t *testing.T) {
	r := newListRig(t)
	u, _ := r.user("after@example.com", "lists:write")
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	bus := events.New(slog.Default(), 0)
	events.Listen(bus, func(_ context.Context, e library.ListItemAdded) error {
		got = append(got, "added:"+e.ItemID)
		return nil
	})
	events.Listen(bus, func(_ context.Context, e library.ListItemRemoved) error {
		got = append(got, "removed:"+e.ItemID)
		return nil
	})

	svc := &library.Lists{DB: db, Events: bus, Now: time.Now}
	l, err := svc.Create(t.Context(), u.ID, library.CreateInput{Type: "game", Title: "Ev", IsPublic: true})
	if err != nil {
		t.Fatal(err)
	}
	game, movie := r.item("game", "public"), r.item("movie", "public")

	// A rejected add (wrong type) rolls back and dispatches nothing.
	if _, err := svc.AddItem(t.Context(), l, movie, ""); err == nil {
		t.Fatal("movie accepted into a game list")
	}
	if len(got) != 0 {
		t.Fatalf("events before commit: %v", got)
	}

	if _, err := svc.AddItem(t.Context(), l, game, ""); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveItem(t.Context(), l, game); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveItem(t.Context(), l, game); err != nil {
		t.Fatal(err)
	}
	if want := "added:" + game + ",removed:" + game; strings.Join(got, ",") != want {
		t.Errorf("events %v, want %s", got, want)
	}
}

// TestMergeMovesListItems: merged items collapse without duplicates and keep the lists consistent.
func TestMergeMovesListItems(t *testing.T) {
	r := newListRig(t)
	u, _ := r.user("merge@example.com", "lists:write")
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	q, err := r.app.Queue()
	if err != nil {
		t.Fatal(err)
	}

	svc := &library.Lists{DB: db, Now: time.Now}
	from, into, other := r.item("game", "public"), r.item("game", "public"), r.item("game", "public")
	both, err := svc.Create(
		t.Context(),
		u.ID,
		library.CreateInput{Type: "game", Title: "Both", IsPublic: true, FirstItemID: from},
	)
	if err != nil {
		t.Fatal(err)
	}
	only, err := svc.Create(
		t.Context(),
		u.ID,
		library.CreateInput{Type: "game", Title: "Only", IsPublic: true, FirstItemID: other},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddItem(t.Context(), only, from, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddItem(t.Context(), both, into, ""); err != nil {
		t.Fatal(err)
	}

	if err := catalog.Merge(t.Context(), db, q, from, into, 0); err != nil {
		t.Fatal(err)
	}

	for _, l := range []struct{ id, items string }{{both.ID, "[2]"}, {only.ID, "[1 2]"}} {
		if got := queryInts(
			t,
			r,
			`SELECT position FROM list_items WHERE list_id = $1 ORDER BY position`,
			l.id,
		); fmt.Sprint(
			got,
		) != l.items {
			t.Errorf("list %s positions = %v, want %s", l.id, got, l.items)
		}
	}
	if got := queryInts(t, r, `SELECT count(*) FROM list_items WHERE item_id = $1`, from); got[0] != 0 {
		t.Errorf("source still referenced by %d entries", got[0])
	}
	if got := queryInts(
		t,
		r,
		`SELECT count(*) FROM list_items WHERE list_id = $1 AND item_id = $2`,
		both.ID,
		into,
	); got[0] != 1 {
		t.Errorf("duplicate target entries: %d", got[0])
	}
	if got := queryInts(t, r, `SELECT item_count FROM lists WHERE id = $1`, both.ID); got[0] != 1 {
		t.Errorf("item_count of the collapsed list = %d, want 1", got[0])
	}
	if got := queryInts(t, r, `SELECT item_count FROM lists WHERE id = $1`, only.ID); got[0] != 2 {
		t.Errorf("item_count of the moved list = %d, want 2", got[0])
	}
}
