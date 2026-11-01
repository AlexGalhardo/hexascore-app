//go:build integration

package tests

import (
	"net/url"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const editItemPath = "/admin/catalog/item/00000000-0000-0000-0000-0000000000aa/fields"

// SDD 04 §7.4: a panel edit writes one catalog_audit row per changed field (source=admin) with the real previous
// value and provenance, in the same transaction as the write.
func TestAdminEditWritesAudit(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedModerationItem(t, app)

	admin := browserFor(t, app, makeUser(t, app, "edit-admin@example.com", models.RoleAdmin, true))
	admin.PostForm(editItemPath, url.Values{
		"field": {"title", "external_ids.imdb", "release_date"},
		"value": {"b", "tt1", ""},
	}).AssertRedirect("/admin/items/00000000-0000-0000-0000-0000000000aa/manage")

	if got := itemTitle(t, app); got != "b" {
		t.Fatalf("title = %q", got)
	}

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.QueryContext(t.Context(), `SELECT field, source::text, actor_role::text, before::text, after::text,
COALESCE(prev_provenance::text, '') FROM catalog_audit WHERE target_id = '00000000-0000-0000-0000-0000000000aa'
ORDER BY field`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	var got [][6]string
	for rows.Next() {
		var r [6]string
		if err := rows.Scan(&r[0], &r[1], &r[2], &r[3], &r[4], &r[5]); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	// release_date was already empty, so it is not a change.
	want := [][6]string{
		{"external_ids.imdb", "admin", "admin", "null", `"tt1"`, "null"},
		{"title", "admin", "admin", `"a"`, `"b"`, `"wikidata"`},
	}
	if len(got) != len(want) {
		t.Fatalf("audit rows = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("row %d = %v, want %v", i, got[i], want[i])
		}
	}

	var prov string
	if err := db.QueryRowContext(t.Context(), `SELECT provenance ->> 'original_title' FROM items
WHERE id = '00000000-0000-0000-0000-0000000000aa'`).Scan(&prov); err != nil || prov != "hexascore" {
		t.Fatalf("provenance = %q, %v", prov, err)
	}

	var changes int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM api_changes
WHERE item_id = '00000000-0000-0000-0000-0000000000aa'`).Scan(&changes); err != nil || changes != 1 {
		t.Fatalf("api_changes = %d, %v", changes, err)
	}

	// An unknown field is refused and writes nothing.
	admin.PostForm(editItemPath, url.Values{"field": {"cover"}, "value": {"x"}}).AssertStatus(303)
}

// Owner Q-A15: catalogue writes are admin-only (applySuggestion), even for a trusted moderator.
func TestSecurity_A01_ModeratorCannotEditCatalog(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	seedModerationItem(t, app)

	for _, role := range []string{models.RoleModerator, "user"} {
		b := browserFor(t, app, makeUser(t, app, "edit-"+role+"@example.com", role, true))
		resp := b.PostForm(editItemPath, url.Values{"field": {"title"}, "value": {"b"}})
		if resp.Status() != 403 && resp.Status() != 404 {
			t.Fatalf("%s: status = %d", role, resp.Status())
		}
	}

	if got := itemTitle(t, app); got != "a" {
		t.Fatalf("non-admin wrote the catalogue: %q", got)
	}
}
