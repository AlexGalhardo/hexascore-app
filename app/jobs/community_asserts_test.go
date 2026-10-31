//go:build integration

package jobs

import (
	"strings"
	"testing"
)

// TestCommunityAsserts runs the SDD 04 §10 nightly asserts on a clean schema, then breaks each one.
func TestCommunityAsserts(t *testing.T) {
	db := openJobsDB(t)
	ctx := t.Context()
	if err := communityAsserts(ctx, db); err != nil {
		t.Fatalf("clean database: %v", err)
	}

	exec(t, db, `INSERT INTO users (id, name, email, password, created_at, updated_at)
VALUES (1, 'A', 'a@x.io', 'x', now(), now())`)
	exec(t, db, `INSERT INTO suggestions (id, author_id, kind, target_type, target_id, target_domain, locale, changes,
  status, submitted_at)
VALUES ('00000000-0000-7000-8000-0000000000c1', 1, 'edit', 'item', $1, 'game', 'en', '{}', 'applied', now())`,
		changedItem)

	for _, step := range []struct{ want, fix string }{
		{"applied without audit rows", `INSERT INTO catalog_audit
(actor_id, actor_role, source, suggestion_id, target_type, target_id, field, before, after)
VALUES (1, 'admin', 'community', '00000000-0000-7000-8000-0000000000c1', 'item', '` + changedItem + `', 'title',
  '"a"', '"b"')`},
		{"stats differ from rebuild", `INSERT INTO contributor_stats (user_id, submitted_total, approved_total)
VALUES (1, 1, 1)`},
	} {
		err := communityAsserts(ctx, db)
		if err == nil || !strings.Contains(err.Error(), step.want) {
			t.Fatalf("want %q, got %v", step.want, err)
		}
		exec(t, db, step.fix)
	}
	if err := communityAsserts(ctx, db); err != nil {
		t.Fatalf("after fixes: %v", err)
	}

	exec(t, db, `INSERT INTO suggestions (id, author_id, kind, target_type, target_id, target_domain, locale, changes,
  status, submitted_at)
VALUES ('00000000-0000-7000-8000-0000000000c2', NULL, 'edit', 'entity', '00000000-0000-0000-0000-0000000000ee',
  'game', 'en', '{}', 'submitted', now())`)
	if err := communityAsserts(ctx, db); err == nil || !strings.Contains(err.Error(), "open on a missing target") {
		t.Fatalf("missing target: %v", err)
	}
}
