//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// ADR 0026: the React resource screens read an allow-list of columns. A non-admin sees 404, request input never
// reaches the SQL text, a resource without a detail page has no /{id} route, and a secret column is never selected.
func TestAdminResourceScreens(t *testing.T) {
	t.Parallel()

	app := catalogAdminApp(t)
	admin := browserFor(t, app, makeUser(t, app, "res-admin@example.com", models.RoleAdmin, true))
	target := makeUser(t, app, "res-target@example.com", "user", true)
	makeUser(t, app, "res-other@example.com", "user", true)

	member := browserFor(t, app, makeUser(t, app, "res-member@example.com", models.RoleModerator, true))
	member.Get("/admin/r/users").AssertStatus(404)
	member.Get("/admin/r/flags").AssertStatus(404)

	rows := inertiaProp(t, admin.Get("/admin/r/users").AssertOK().Body(), "Admin/ResourceList", "rows")
	for _, want := range []string{"res-admin@example.com", "res-target@example.com"} {
		if !strings.Contains(rows, want) {
			t.Fatalf("users list misses %q: %s", want, rows)
		}
	}

	// Search is bound and escaped: a prefix finds one user, SQL and LIKE metacharacters find none and do not fail.
	one := inertiaProp(t, admin.Get("/admin/r/users?q=res-target").AssertOK().Body(), "Admin/ResourceList", "rows")
	if !strings.Contains(one, "res-target@example.com") || strings.Contains(one, "res-other@example.com") {
		t.Fatalf("search did not narrow the list: %s", one)
	}

	for _, q := range []string{"'; DROP TABLE users; --", "%", "_"} {
		got := inertiaProp(
			t,
			admin.Get("/admin/r/users?q="+url.QueryEscape(q)).AssertOK().Body(),
			"Admin/ResourceList",
			"rows",
		)
		if strings.Contains(got, "@example.com") {
			t.Fatalf("q=%q matched users: %s", q, got)
		}
	}

	// An unknown sort key or filter value falls back to the defaults instead of reaching the SQL.
	admin.Get("/admin/r/users?sort=id;DROP+TABLE+users&dir=asc&f_role=admin'+OR+1=1").AssertOK()

	// The detail page lists whitelisted fields only.
	detail := inertiaProp(
		t,
		admin.Get("/admin/r/users/"+itoa(target.ID)).AssertOK().Body(),
		"Admin/ResourceShow",
		"cells",
	)
	if !strings.Contains(detail, "res-target@example.com") {
		t.Fatalf("detail misses the email: %s", detail)
	}

	for _, secret := range []string{"password", "two_factor_secret", "two_factor_recovery_codes", "stripe_customer_id", "unused"} {
		if strings.Contains(detail, secret) || strings.Contains(rows, secret) {
			t.Fatalf("a secret column %q reached the page", secret)
		}
	}

	admin.Get("/admin/r/users/999999").AssertNotFound()
	admin.Get("/admin/r/nope").AssertNotFound()
	admin.Get("/admin/r/rating-flags/1").AssertNotFound() // no detail page for this resource
	admin.Get("/admin/r/flags").AssertOK()

	// Every screen's list query runs against the real schema, empty or not.
	for _, slug := range []string{
		"users", "score-bursts", "rating-flags", "score-actions", "score-formula-versions", "reports", "contact-messages",
		"items", "job-runs", "ingest-failures", "dq-results", "audit-logs",
	} {
		admin.Get("/admin/r/" + slug).AssertOK()
		admin.Get("/admin/r/" + slug + "?q=x&sort=id&dir=asc&page=2").AssertOK()
	}
}
