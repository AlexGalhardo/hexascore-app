//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/forms"
)

func contactToken(t *testing.T, r *listRig) string {
	t.Helper()

	key, err := r.app.DeriveKey("contact.form", 32)
	if err != nil {
		t.Fatal(err)
	}

	return forms.Token{Key: key, Now: r.app.Now}.IssueAt("contact", r.app.Now().Add(-forms.MinFill))
}

// A signed-in sender cannot speak for someone else: the stored name and email come from the session account
// (the nickname when it has one), whatever the form posts.
func TestSecurity_A01_ContactUsesSessionIdentity(t *testing.T) {
	r := newListRig(t)
	u := makeUser(t, r.app, "contact-me@example.com", "user", true)

	if err := r.acc.SetNickname(t.Context(), u.ID, "contactme", "nickname"); err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"name": {"Someone Else"}, "email": {"victim@example.com"}, "subject": {"question"},
		"message": {"this message is long enough to pass the thirty-two minimum"}, "token": {contactToken(t, r)},
	}
	browserFor(t, r.app, u).PostForm("/contact", form).AssertStatus(303)

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var name, email string
	if err := db.QueryRowContext(t.Context(), `SELECT name, email FROM contact_messages WHERE user_id = $1`, u.ID).
		Scan(&name, &email); err != nil {
		t.Fatal(err)
	}

	if name != "contactme" || email != "contact-me@example.com" {
		t.Fatalf("stored %q (%s), want the session account", name, email)
	}
}

func TestContactMessageBounds(t *testing.T) {
	r := newListRig(t)
	c := r.app.Client(t).WithHeader("Sec-Fetch-Site", "same-origin").WithHeader("Accept", "application/json")
	c.PostForm("/contact", url.Values{
		"name": {"Ana"}, "email": {"ana@example.com"}, "subject": {"bug"},
		"message": {strings.Repeat("a", 512)}, "token": {contactToken(t, r)},
	})
	if n := r.count(`SELECT count(*) FROM contact_messages`); n != 1 {
		t.Fatalf("512 characters: %d stored, want 1", n)
	}

	c.PostForm("/contact", url.Values{
		"name": {"Ana"}, "email": {"ana@example.com"}, "subject": {"bug"},
		"message": {strings.Repeat("a", 513)}, "token": {contactToken(t, r)},
	}).AssertUnprocessable()
}
