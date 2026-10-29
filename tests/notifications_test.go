//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/notifications"
)

func sendNote(t *testing.T, r *listRig, userID int64, key string) bool {
	t.Helper()

	created, err := notifications.Send(
		t.Context(),
		r.acc.DB,
		userID,
		notifications.System,
		key,
		map[string]any{"k": key},
	)
	if err != nil {
		t.Fatal(err)
	}

	return created
}

func TestNotificationDedupeOnRetry(t *testing.T) {
	r := newListRig(t)
	u, _ := r.user("dedupe@example.com", "catalog:read")
	other, _ := r.user("dedupe2@example.com", "catalog:read")

	if !sendNote(t, r, u.ID, "report:1") {
		t.Fatal("first send must insert")
	}

	// A retried job sends the same key again.
	if sendNote(t, r, u.ID, "report:1") {
		t.Fatal("retry must not insert")
	}

	sendNote(t, r, u.ID, "report:2")
	// The key is scoped to the user.
	sendNote(t, r, other.ID, "report:1")

	n, err := notifications.Unread(t.Context(), r.acc.DB, u.ID)
	if err != nil || n != 2 {
		t.Fatalf("unread = %d, %v; want 2", n, err)
	}
}

func TestSecurity_A01_NotificationsOwnerOnly(t *testing.T) {
	r := newListRig(t)
	owner, ownerAPI := r.user("noteowner@example.com", "catalog:read", "profile:write")
	stranger, strangerAPI := r.user("notestranger@example.com", "catalog:read", "profile:write")
	_, readOnly := r.user("noteread@example.com", "catalog:read")

	sendNote(t, r, owner.ID, "a")
	sendNote(t, r, owner.ID, "b")
	sendNote(t, r, stranger.ID, "c")

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Unread int `json:"unread"`
		} `json:"meta"`
	}

	ownerAPI.GetJSON("/app/v1/me/notifications").AssertOK().JSON(&list)
	if len(list.Data) != 2 || list.Meta.Unread != 2 {
		t.Fatalf("owner list = %+v", list)
	}

	// Someone else's id answers like an unknown one, and changes nothing.
	id := list.Data[0].ID
	strangerAPI.PostJSON("/app/v1/me/notifications/read", map[string]any{"ids": []string{id}}).AssertNotFound()
	strangerAPI.PostJSON("/app/v1/me/notifications/read", map[string]any{"ids": []string{uuid.NewV7().String()}}).
		AssertNotFound()
	strangerAPI.PostJSON("/app/v1/me/notifications/read", map[string]any{"ids": []string{"nope"}}).AssertNotFound()
	readOnly.PostJSON("/app/v1/me/notifications/read", map[string]any{"all": true}).AssertForbidden()

	// Nothing changed for the owner.
	strangerAPI.PostJSON("/app/v1/me/notifications/read", map[string]any{"ids": []string{id}}).AssertNotFound()
	n, _ := notifications.Unread(t.Context(), r.acc.DB, owner.ID)
	if n != 2 {
		t.Fatalf("owner unread = %d, want 2", n)
	}

	r.app.Client(t).GetJSON("/app/v1/me/notifications").AssertUnauthorized()

	ownerAPI.PostJSON("/app/v1/me/notifications/read", map[string]any{"ids": []string{id}}).AssertNoContent()
	ownerAPI.GetJSON("/app/v1/me/notifications?unread=1").AssertOK().JSON(&list)
	if len(list.Data) != 1 || list.Meta.Unread != 1 {
		t.Fatalf("after read = %+v", list)
	}

	// The same through the session guard: page, shared badge, mark read.
	web := browserFor(t, r.app, owner)
	web.Get("/me/notifications").AssertOK().AssertInertia("Account/Notifications").
		HasProp("notifications.data", 2).PropEquals("unreadNotifications", 1)

	strangerWeb := browserFor(t, r.app, stranger)
	strangerWeb.PostJSON("/me/notifications/"+id+"/read", nil).AssertNotFound()
	strangerWeb.Get("/me/notifications").AssertInertia("Account/Notifications").
		HasProp("notifications.data", 1).PropEquals("unreadNotifications", 1)

	web.PostJSON("/me/notifications/read-all", nil).AssertRedirect("/me/notifications")
	web.Get("/me/notifications").AssertInertia("Account/Notifications").PropEquals("unreadNotifications", 0)
	strangerWeb.Get("/me/notifications").AssertInertia("Account/Notifications").PropEquals("unreadNotifications", 1)
}
