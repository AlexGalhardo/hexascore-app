//go:build integration

package tests

import (
	"fmt"
	"net/http"
	"testing"
)

// TestAppLibrary covers the library and public profile routes that already
// exist on /app/v1. A private profile returns the nickname only.
func TestAppLibrary(t *testing.T) {
	t.Parallel()

	r := newListRig(t)
	owner, mine := r.user("app-lib@example.com", "catalog:read")
	_, other := r.user("app-lib-other@example.com", "catalog:read")
	r.setNickname(owner, "privlib", true)

	mine.GetJSON("/app/v1/me/library").AssertOK()
	var body map[string]any
	other.GetJSON("/app/v1/users/privlib").AssertOK().JSON(&body)
	if _, ok := body["stats"]; ok {
		t.Fatal("private profile included stats")
	}
	if _, ok := body["ratings"]; ok {
		t.Fatal("private profile included ratings")
	}
	if _, ok := body["lists"]; ok {
		t.Fatal("private profile included lists")
	}

	profile, _ := body["profile"].(map[string]any)
	if profile["nickname"] != "privlib" || profile["isPrivate"] != true {
		t.Fatalf("profile = %#v", profile)
	}
}

// TestAppNotifications: more than 100 ids is 400. One hundred is inside the cap.
func TestAppNotifications(t *testing.T) {
	t.Parallel()

	r := newListRig(t)
	_, c := r.user("app-notes@example.com", "profile:write")
	ids := make([]string, 101)
	for i := range ids {
		ids[i] = fmt.Sprintf("id-%03d", i)
	}

	c.PostJSON("/app/v1/me/notifications/read", map[string]any{"ids": ids}).AssertStatus(http.StatusBadRequest)
	c.PostJSON("/app/v1/me/notifications/read", map[string]any{"ids": ids[:100]}).AssertNotFound()
}

// TestAppReports: reporting yourself is 403 report-self.
func TestAppReports(t *testing.T) {
	t.Parallel()

	r := newListRig(t)
	owner, c := r.user("app-rep@example.com", "lists:write", "profile:write")
	r.setNickname(owner, "apprep", false)
	c.PostJSON("/app/v1/reports", reportBody("user", "apprep")).AssertForbidden().AssertSee("report-self")
}
