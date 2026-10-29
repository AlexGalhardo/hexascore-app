//go:build integration

package scoring_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

func TestStatus(t *testing.T) {
	r := newRig(t)
	c := r.token(1, "status:write", "ratings:write")
	item := r.item("game", "2020-01-01", "day", "released")
	path := "/app/v1/items/" + item + "/status"

	c.PutJSON(path, map[string]any{"status": "bogus"}).AssertUnprocessable()
	c.PutJSON(path, map[string]any{}).AssertUnprocessable()
	r.token(9, "ratings:write").PutJSON(path, map[string]any{"status": "completed"}).AssertForbidden()
	c.PutJSON("/app/v1/items/not-a-uuid/status", map[string]any{"status": "completed"}).AssertNotFound()

	for _, s := range []string{"want", "in_progress", "completed", "dropped"} {
		c.PutJSON(path, map[string]any{"status": s}).AssertOK().AssertJSONPath("status", s)
	}
	vukemtest.AssertDatabaseHas(t, r.app, "item_status", map[string]any{"item_id": item, "status": "dropped"})
	vukemtest.AssertDatabaseHas(t, r.app, "rating_events", map[string]any{"item_id": item, "kind": "status_changed"})

	c.Delete(path).AssertNoContent()
	vukemtest.AssertDatabaseMissing(t, r.app, "item_status", map[string]any{"item_id": item})
}

func TestTimeToBeat(t *testing.T) {
	r := newRig(t)
	c := r.token(1, "status:write")
	game := r.item("game", "2020-01-01", "day", "released")
	movie := r.item("movie", "2020-01-01", "day", "released")
	ttb := "/app/v1/items/" + game + "/time-to-beat"
	body := func(m any) map[string]any { return map[string]any{"mainHours": m} }

	c.PutJSON(ttb, body(10.0)).AssertStatus(409).AssertJSONPath("title", "Conflict")

	c.PutJSON("/app/v1/items/"+movie+"/status", map[string]any{"status": "completed"}).AssertOK()
	c.PutJSON("/app/v1/items/"+movie+"/time-to-beat", body(10.0)).
		AssertUnprocessable().
		AssertJSONPath("title", "Unprocessable Entity")

	c.PutJSON("/app/v1/items/"+game+"/status", map[string]any{"status": "completed"}).AssertOK()
	for _, bad := range []any{0.25, 0.0, 10000.5, 12.3, "x"} {
		c.PutJSON(ttb, body(bad)).AssertUnprocessable()
	}
	c.PutJSON(ttb, map[string]any{}).AssertUnprocessable()
	c.PutJSON(ttb, map[string]any{"mainHours": nil, "extrasHours": nil}).AssertUnprocessable()

	c.PutJSON(ttb, map[string]any{"mainHours": 12.5, "extrasHours": 0.5, "completionistHours": 10000.0}).AssertOK()
	vukemtest.AssertDatabaseHas(t, r.app, "time_to_beat",
		map[string]any{"item_id": game, "main_minutes": 750, "extras_minutes": 30, "completionist_minutes": 600000})
	me := c.Get("/app/v1/items/" + game + "/me").AssertOK()
	me.AssertJSONPath("status", "completed")
	me.AssertJSONPath("timeToBeat.mainHours", 12.5)
	me.AssertJSONPath("rating", nil)

	// Leaving "completed" drops the report.
	c.PutJSON("/app/v1/items/"+game+"/status", map[string]any{"status": "dropped"}).AssertOK()
	vukemtest.AssertDatabaseMissing(t, r.app, "time_to_beat", map[string]any{"item_id": game})
	vukemtest.AssertDatabaseHas(
		t,
		r.app,
		"rating_events",
		map[string]any{"item_id": game, "kind": "time_to_beat_cleared"},
	)

	c.Delete(ttb).AssertNoContent()
	c.Get("/app/v1/items/00000000-0000-7000-8000-000000000000/me").AssertNotFound()
}

func TestStatusAndTimeToBeatEvents(t *testing.T) {
	r := newRig(t)
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	var uid int64
	if err := db.SQL().QueryRowContext(t.Context(), `INSERT INTO users (name, email, password, created_at, updated_at)
VALUES ('E', 'ev@example.com', 'x', now(), now()) RETURNING id`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	item := r.item("game", "2020-01-01", "day", "released")

	bus := events.New(slog.Default(), 0)
	var mu sync.Mutex
	var got []string
	record := func(name string) func() error {
		return func() error { mu.Lock(); defer mu.Unlock(); got = append(got, name); return nil }
	}
	events.Listen(bus, func(_ context.Context, _ scoring.StatusChanged) error { return record("status")() })
	events.Listen(bus, func(_ context.Context, _ scoring.TimeToBeatReported) error { return record("ttb")() })
	events.Listen(bus, func(_ context.Context, _ scoring.ItemScoreChanged) error { return record("score")() })

	w := &scoring.Writer{DB: db, Events: bus, Now: time.Now}
	hours := 8.0
	if _, err := w.SetStatus(t.Context(), uid, item, "completed"); err != nil {
		t.Fatal(err)
	}
	if _, err := w.SaveTimeToBeat(t.Context(), uid, item, scoring.TimeToBeatHours{Main: &hours}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.SaveTimeToBeat(t.Context(), uid, item, scoring.TimeToBeatHours{Main: &hours}); err != nil {
		t.Fatal(err)
	}
	// A rejected write dispatches nothing.
	if _, err := w.SetStatus(t.Context(), uid, "00000000-0000-7000-8000-000000000000", "completed"); err == nil {
		t.Fatal("unknown item accepted")
	}

	mu.Lock()
	defer mu.Unlock()
	want := []string{"status", "score", "ttb", "score", "ttb", "score"}
	if len(got) != len(want) {
		t.Fatalf("events %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events %v, want %v", got, want)
		}
	}
}
