//go:build integration

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/library"
)

func TestLibraryCommunityScore(t *testing.T) {
	r := newListRig(t)
	u, api := r.user("libscore@example.com", "catalog:read", "status:write")
	r.setNickname(u, "ScoreFan", false)

	scored, gated := r.item("game", "public"), r.item("game", "public")
	r.exec(`UPDATE items SET rating_count = 7, score = 84 WHERE id = $1`, scored)
	r.exec(`UPDATE items SET rating_count = 3, score = 70 WHERE id = $1`, gated)

	for _, id := range []string{scored, gated} {
		r.status(u.ID, id, "completed")
		r.exec(
			`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 90, 90, 90, 90, 90, 90)`,
			u.ID,
			id,
		)
	}

	want := map[string]string{scored: "8.4", gated: "<nil>"}
	check := func(t *testing.T, where, id string, present bool, score *float64) {
		t.Helper()

		if !present {
			t.Fatalf("%s: score field missing for %s", where, id)
		}

		got := "<nil>"
		if score != nil {
			got = fmt.Sprint(*score)
		}

		if got != want[id] {
			t.Fatalf("%s: item %s score = %s, want %s", where, id, got, want[id])
		}
	}

	t.Run("library", func(t *testing.T) {
		var view struct {
			Groups []struct {
				Items []map[string]any `json:"items"`
			} `json:"groups"`
		}
		api.GetJSON("/app/v1/me/library").AssertOK().JSON(&view)

		for _, it := range view.Groups[0].Items {
			v, ok := it["score"]
			var s *float64
			if f, isF := v.(float64); isF {
				s = &f
			}

			check(t, "library", it["id"].(string), ok, s)
		}
	})

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("profile", func(t *testing.T) {
		p, err := library.LoadProfile(t.Context(), db, "scorefan", u.ID, "en", 0)
		if err != nil {
			t.Fatal(err)
		}

		if len(p.Ratings) != 2 {
			t.Fatalf("ratings = %d", len(p.Ratings))
		}

		for _, rt := range p.Ratings {
			check(t, "profile", rt.ID, true, rt.Score)
		}
	})

	t.Run("list entries", func(t *testing.T) {
		svc := &library.Lists{DB: db, Now: time.Now}
		l, err := svc.Create(t.Context(), u.ID,
			library.CreateInput{Type: "game", Title: "L", IsPublic: true, FirstItemID: scored})
		if err != nil {
			t.Fatal(err)
		}

		if _, err := db.ExecContext(t.Context(),
			`INSERT INTO list_items (list_id, item_id, position) VALUES ($1, $2, 2)`, l.ID, gated); err != nil {
			t.Fatal(err)
		}

		entries, err := svc.Entries(t.Context(), l.ID, "en")
		if err != nil {
			t.Fatal(err)
		}

		if len(entries) != 2 {
			t.Fatalf("entries = %d", len(entries))
		}

		for _, e := range entries {
			check(t, "list", e.Item.ID, true, e.Item.Score)
		}
	})
}
