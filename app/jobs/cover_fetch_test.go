//go:build integration

package jobs

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// TestCoverFetch_RetryUploadsOnce runs the handler twice on one payload: the second run sees the stored cover_status
// and neither downloads nor stores again.
func TestCoverFetch_RetryUploadsOnce(t *testing.T) {
	db := openJobsDB(t)
	const sid = "00000000-0000-7000-8000-0000000000f1"
	exec(t, db, `INSERT INTO suggestions (id, author_id, kind, target_type, target_id, target_domain, locale, changes,
  status, submitted_at)
VALUES ($1, NULL, 'edit', 'item', $2, 'game', 'en', '{}', 'submitted', now())`, sid, changedItem)

	calls := 0
	fetch := func(context.Context, storage.Disk, string) (community.Cover, error) {
		calls++

		return community.Cover{Key: "community/covers/a.png", Width: 400, Height: 400}, nil
	}
	job := CoverFetch{SuggestionID: sid, URL: "https://example.com/c.png"}
	for range 2 {
		if err := runCoverFetch(t.Context(), db, nil, fetch, job); err != nil {
			t.Fatal(err)
		}
	}

	if calls != 1 {
		t.Fatalf("fetches = %d, want 1", calls)
	}
}
