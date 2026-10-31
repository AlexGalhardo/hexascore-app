//go:build integration

package tests

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
	"github.com/AlexGalhardo/hexascore-app/app/jobs"
)

func TestAPIChanges(t *testing.T) {
	clock := vukemtest.NewFakeClock(time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC))
	// The free plan's 1 rps throttle is not under test here (same fake as newAPICatalog).
	r := newKeyRigWith(t,
		vukem.WithClock(clock.Now),
		vukem.WithRateLimiters(func(_ string, _ int, window time.Duration) ratelimit.Limiter {
			return ratelimit.NewMemory(1000, window)
		}),
	)
	client := r.app.Client(t).WithToken(r.create(r.signIn("changes@example.com"), "ci"))
	now := clock.Now()

	game := uuid.NewV7().String()
	movie := uuid.NewV7().String()
	target := uuid.NewV7().String()
	insertChange(t, r, game, "game", "updated", "{title,score}", "", now.Add(-10*24*time.Hour))
	insertChange(t, r, movie, "movie", "deleted", "{}", target, now)
	kept := insertChange(t, r, uuid.NewV7().String(), "book", "created", "{}", "", now.Add(-changes.Retain))
	gone := insertChange(
		t,
		r,
		uuid.NewV7().String(),
		"series",
		"updated",
		"{slug}",
		"",
		now.Add(-changes.Retain-time.Minute),
	)

	base := "/api/v1/changes?" + url.Values{"since": {now.Add(-changes.Window).Format(time.RFC3339)}}.Encode()

	t.Run("window", func(*testing.T) {
		tooOld := "/api/v1/changes?" + url.Values{
			"since": {now.Add(-changes.Window - time.Second).Format(time.RFC3339)},
		}.Encode()
		client.GetJSON(tooOld).AssertStatus(http.StatusBadRequest).
			AssertJSONPath("code", "changes-window-exceeded")

		future := "/api/v1/changes?" + url.Values{"since": {now.Add(time.Second).Format(time.RFC3339)}}.Encode()
		client.GetJSON(future).AssertStatus(http.StatusBadRequest).AssertJSONPath("code", "invalid-query")
		client.GetJSON("/api/v1/changes").AssertStatus(http.StatusBadRequest).AssertJSONPath("code", "invalid-query")
		client.GetJSON(base+"&limit=1001").AssertStatus(http.StatusBadRequest).AssertJSONPath("code", "invalid-query")

		client.GetJSON(base).AssertOK().AssertMatchesOpenAPI().
			AssertJSONPath("data.0.itemId", game).
			AssertJSONPath("data.0.action", "updated").
			AssertJSONPath("data.0.itemType", "games").
			AssertJSONPath("data.0.fields", []string{"title", "score"}).
			AssertJSONPath("data.1.action", "deleted").
			AssertJSONPath("data.1.itemType", "movies").
			AssertJSONPath("data.1.itemId", movie).
			AssertJSONPath("data.1.mergedInto", target)
	})

	t.Run("cursor and type", func(t *testing.T) {
		client.GetJSON(base+"&type=movies").AssertOK().AssertMatchesOpenAPI().
			AssertJSONPath("data.0.action", "deleted").
			AssertJSONPath("data.0.mergedInto", target)
		client.GetJSON(base + "&type=films").AssertStatus(http.StatusBadRequest)
		client.GetJSON(base+"&cursor=garbage").AssertStatus(http.StatusBadRequest).
			AssertJSONPath("code", "invalid-cursor")

		page := client.GetJSON(base+"&limit=1").AssertOK().AssertJSONPath("data.0.itemId", game)
		var body struct {
			Meta struct {
				NextCursor string `json:"next_cursor"`
			} `json:"meta"`
		}
		page.JSON(&body)
		if body.Meta.NextCursor == "" {
			t.Fatal("next_cursor empty")
		}

		next := base + "&limit=1&cursor=" + url.QueryEscape(body.Meta.NextCursor)
		client.GetJSON(next).AssertOK().
			AssertJSONPath("data.0.action", "deleted").
			AssertJSONPath("data.0.mergedInto", target)
	})

	t.Run("prune", func(t *testing.T) {
		n, err := changes.Prune(t.Context(), r.acc.DB, now)
		if err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("pruned %d, want 1", n)
		}

		var still, absent int
		if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM api_changes WHERE id = $1`, kept).
			Scan(&still); err != nil {
			t.Fatal(err)
		}
		if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM api_changes WHERE id = $1`, gone).
			Scan(&absent); err != nil {
			t.Fatal(err)
		}
		if still != 1 || absent != 0 {
			t.Fatalf("kept = %d, gone = %d, want the row at exactly 30 days to stay", still, absent)
		}

		sched, err := r.app.Schedule()
		if err != nil {
			t.Fatal(err)
		}
		want := "job:" + jobs.PruneAPIChanges{}.Kind()
		for _, task := range sched.Tasks() {
			if task.Name() != want {
				continue
			}
			if task.Expression() != "50 4 * * *" {
				t.Fatalf("prune cron = %q, want 50 4 * * *", task.Expression())
			}

			return
		}
		t.Fatalf("schedule missing %s", want)
	})
}

func insertChange(t *testing.T, r *keyRig, itemID, itemType, action, fields, merged string, at time.Time) int64 {
	t.Helper()

	var mergedArg any
	if merged != "" {
		mergedArg = merged
	}

	var id int64
	err := r.acc.DB.QueryRowContext(t.Context(), `
INSERT INTO api_changes (item_id, item_type, action, fields, merged_into, changed_at)
VALUES ($1, $2, $3, $4::text[], $5, $6)
RETURNING id`, itemID, itemType, action, fields, mergedArg, at).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	return id
}
