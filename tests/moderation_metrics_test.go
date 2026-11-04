//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func TestModerationMetricsPage(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	mod := browserFor(t, app, makeUser(t, app, "m-mod@example.com", models.RoleModerator, true))
	got, props := inertiaPage(t, mod.Get("/admin/moderation/metrics").AssertOK().Body())
	if got != "Admin/ModerationMetrics" || props["week"] == nil || props["month"] == nil {
		t.Fatalf("metrics page = %q, props %v", got, props)
	}
}

// The metrics page reads the moderation queue, so it needs the moderate ability like the queue itself.
func TestSecurity_A01_ModerationMetricsNeedsAbility(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	user := browserFor(t, app, makeUser(t, app, "m-user@example.com", "user", true))
	user.Get("/admin/moderation/metrics").AssertForbidden()

	mod := browserFor(t, app, makeUser(t, app, "m-untrusted@example.com", models.RoleModerator, true))
	mod.Get("/admin/moderation/metrics").AssertForbidden()
}
