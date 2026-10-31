//go:build integration

package tests

import (
	"net/url"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
)

func ownSuggestion(t *testing.T, app *vukemtest.App, id, itemID, domain string, author int64) {
	t.Helper()
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, submitted_at)
VALUES ($1, $2, 'edit', 'item', $3, $4, 'en', '{}', 'submitted', now())`, id, author, itemID, domain); err != nil {
		t.Fatal(err)
	}
}

func systemStatus(t *testing.T, app *vukemtest.App, id string) (string, *int64) {
	t.Helper()
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	var status string
	var author *int64
	if err := db.QueryRowContext(t.Context(), `SELECT status::text, author_id FROM suggestions WHERE id = $1`, id).
		Scan(&status, &author); err != nil {
		t.Fatal(err)
	}

	return status, author
}

// The SDD 04 section 8 system transitions run from the real triggers: the admin ban, the deletion purge and an
// admin catalogue merge.
func TestCommunitySystemTransitions(t *testing.T) {
	t.Parallel()

	app := catalogAdminApp(t)
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	admin := browserFor(t, app, makeUser(t, app, "sys-admin@example.com", "admin", true))
	banned := makeUser(t, app, "sys-banned@example.com", "user", true)
	gone := makeUser(t, app, "sys-gone@example.com", "user", true)
	author := makeUser(t, app, "sys-author@example.com", "user", true)

	const onBan, onGone, onMerge = "00000000-0000-7000-8000-0000000000f1", "00000000-0000-7000-8000-0000000000f2",
		"00000000-0000-7000-8000-0000000000f3"
	ownSuggestion(t, app, onBan, itemB, "game", banned.ID)
	ownSuggestion(t, app, onGone, itemB, "game", gone.ID)
	ownSuggestion(t, app, onMerge, itemA, "game", author.ID)

	admin.PostForm("/admin/users/"+itoa(banned.ID)+"/ban", url.Values{"reason": {"spam"}}).
		AssertRedirect("/admin/users/" + itoa(banned.ID))
	if status, _ := systemStatus(t, app, onBan); status != "withdrawn" {
		t.Fatalf("banned author's suggestion = %s", status)
	}

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `UPDATE users SET deletion_scheduled_at = $2 WHERE id = $1`, gone.ID,
		time.Now().UTC().AddDate(0, 0, -31)); err != nil {
		t.Fatal(err)
	}
	if done, err := acc.PurgeDue(t.Context()); err != nil || len(done) != 1 {
		t.Fatalf("purged %v, %v", done, err)
	}
	if status, a := systemStatus(t, app, onGone); status != "withdrawn" || a != nil {
		t.Fatalf("deleted author's suggestion = %s author %v", status, a)
	}

	admin.PostForm("/admin/items/"+itemA+"/merge", url.Values{"into": {itemB}, "confirm": {"merge"}}).
		AssertRedirect("/admin/items/" + itemB)
	if status, _ := systemStatus(t, app, onMerge); status != "expired" {
		t.Fatalf("suggestion on the merged-away item = %s", status)
	}
}
