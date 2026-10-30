//go:build integration

package tests

import (
	"fmt"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/library"
)

// TestAppLists hits the /app/v1 list writes that already exist and the list
// reads added for P3.C.3.2. A stale reorder is 409 list-version-conflict, a
// partial order is 422 order-mismatch, an edition of another work is 422
// edition-not-of-work, and another user's private list is 404.
func TestAppLists(t *testing.T) {
	t.Parallel()

	r := newListRig(t)
	_, owner := r.user("app-lists@example.com", "lists:write", "catalog:read", "status:write")
	_, other := r.user("app-lists-other@example.com", "catalog:read")
	a, b := r.item("game", "public"), r.item("game", "public")
	_, path := createList(t, owner, "game")

	owner.PutJSON(path+"/items/"+a, map[string]any{}).AssertOK()
	owner.PutJSON(path+"/items/"+b, map[string]any{}).AssertOK().AssertJSONPath("version", float64(2))
	owner.PutJSON(path+"/order", map[string]any{"version": 1, "itemIds": []string{b, a}}).
		AssertStatus(409).
		AssertSee(library.CodeVersionConflict)
	owner.PutJSON(path+"/order", map[string]any{"version": 2, "itemIds": []string{a}}).
		AssertUnprocessable().
		AssertSee(library.CodeOrderMismatch)

	book, otherBook := r.item("book", "public"), r.item("book", "public")
	edition := r.edition(otherBook)
	owner.PutJSON("/app/v1/items/"+book+"/edition", map[string]any{"editionId": edition}).
		AssertUnprocessable().
		AssertSee(library.CodeEditionNotOfWork)

	secret, _ := owner.PostJSON("/app/v1/lists", map[string]any{
		"type": "game", "title": "Secret", "isPublic": false,
	}).AssertStatus(201).JSONPath("id")
	id := fmt.Sprint(secret)
	owner.GetJSON("/app/v1/lists/"+id).AssertOK().AssertJSONPath("id", id)
	owner.GetJSON("/app/v1/me/lists").AssertOK().AssertSee(id)
	other.GetJSON("/app/v1/lists/" + id).AssertNotFound()
}
