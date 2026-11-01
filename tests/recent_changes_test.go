//go:build integration

package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// SDD 04 FR-04-13, §7.6, NFR-04-5: the public Recent changes tab lists the target's audit rows, 20 per page, with the
// author's nickname, "HexaScore team" for admins and "deleted user" once the actor id is gone.
func TestRecentChanges(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedModerationItem(t, app)

	author := makeUser(t, app, "rc-author@example.com", "user", true)
	admin := makeUser(t, app, "rc-admin@example.com", models.RoleAdmin, true)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE users SET username = 'rcauthor' WHERE id = $1`,
		author.ID,
	); err != nil {
		t.Fatal(err)
	}

	const item = "00000000-0000-0000-0000-0000000000aa"
	insert := func(actor any, role, source, field, after string) {
		t.Helper()
		if _, err := db.ExecContext(t.Context(), `INSERT INTO catalog_audit
(actor_id, actor_role, source, target_type, target_id, field, before, after)
VALUES ($1, $2, $3, 'item', $4, $5, '"a"', $6::jsonb)`, actor, role, source, item, field, after); err != nil {
			t.Fatal(err)
		}
	}

	for i := range 21 {
		insert(nil, "system", "merge", "filler", fmt.Sprintf(`"%d"`, i))
	}
	insert(author.ID, "author", "community", "title", `"<b>x</b>"`)
	insert(admin.ID, "admin", "admin", "release_date", `{"value":"2001-02-03","precision":"day"}`)
	insert(nil, "author", "community", "external_ids.imdb", `"tt1"`)

	page := app.Client(t).Inertia("").Get("/movies/old/changes").AssertOK().AssertInertia("Catalog/RecentChanges")
	page.HasProp("entries", 20).
		PropEquals("entries.0.actor", "deleted").
		PropEquals("entries.0.field", "external_ids.imdb").
		PropEquals("entries.1.actor", "team").
		PropEquals("entries.2.actor", "user").
		PropEquals("entries.2.nickname", "rcauthor").
		PropEquals("entries.2.after", `"<b>x</b>"`).
		PropEquals("target.title", "a").
		PropEquals("nextPage", float64(2))

	if id, ok := page.Prop("entries.1.nickname"); ok && id != "" && id != nil {
		t.Fatalf("admin nickname leaked: %v", id)
	}

	app.Client(t).Inertia("").Get("/movies/old/changes?page=2").AssertOK().AssertInertia("Catalog/RecentChanges").
		HasProp("entries", 4).PropEquals("nextPage", nil)

	// SSR: the first response is HTML with the entries in the page data.
	html := app.Client(t).Get("/movies/old/changes").AssertOK().Body()
	if !strings.Contains(html, "rcauthor") {
		t.Fatalf("SSR body lacks the entries: %.300s", html)
	}

	app.Client(t).Get("/movies/missing/changes").AssertStatus(301).AssertHeader("Location", "/")
}

// FR-04-15: profiles show contribution counts and approval rate after at least one decision; tier and accuracy only
// to the user (and admins).
func TestProfileContributions(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	u := makeUser(t, app, "pc-user@example.com", "user", true)
	other := makeUser(t, app, "pc-other@example.com", "user", true)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	for q, args := range map[string][]any{
		`UPDATE users SET username = $2 WHERE id = $1`: {u.ID, "pcuser"},
		`INSERT INTO contributor_stats (user_id, submitted_total, approved_total, rejected_total)
VALUES ($1, 5, 3, 1)`: {u.ID},
		`UPDATE users SET username = 'pcother' WHERE id = $1`: {other.ID},
	} {
		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatal(err)
		}
	}

	app.Client(t).Inertia("").Get("/u/pcuser").AssertOK().AssertInertia("User/Profile").
		PropEquals("contributions.submitted", float64(5)).
		PropEquals("contributions.approved", float64(3)).
		PropEquals("contributions.rejected", float64(1)).
		PropEquals("contributions.approvalRate", 0.75).
		MissingProp("contributions.tier")

	browserFor(t, app, u).Inertia("").Get("/u/pcuser").AssertOK().AssertInertia("User/Profile").
		PropEquals("contributions.tier", "new").
		HasProp("contributions.accuracy")

	// No decision yet: no contributions block.
	app.Client(t).Inertia("").Get("/u/pcother").AssertOK().AssertInertia("User/Profile").
		MissingProp("contributions")
}
