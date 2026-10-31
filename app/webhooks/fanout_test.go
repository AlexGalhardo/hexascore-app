package webhooks

import (
	"testing"
	"time"
)

func TestCoalesceKeepsOneEventPerItem(t *testing.T) {
	at := time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC)
	got := coalesce([]changeRow{
		{id: 1, itemID: "game-1", itemType: "game", action: "updated", fields: []string{"score"}, at: at},
		{
			id:       2,
			itemID:   "game-1",
			itemType: "game",
			action:   "updated",
			fields:   []string{"ratingCount", "score"},
			at:       at.Add(time.Second),
		},
		{id: 3, itemID: "movie-1", itemType: "movie", action: "created", fields: []string{"title"}, at: at},
		{
			id:       4,
			itemID:   "movie-1",
			itemType: "movie",
			action:   "deleted",
			merged:   "movie-2",
			at:       at.Add(2 * time.Second),
		},
		{id: 5, itemID: "skip", itemType: "person", action: "updated", fields: []string{"name"}, at: at},
	})

	if len(got) != 2 {
		t.Fatalf("events = %d, want 2 (unknown item type dropped)", len(got))
	}

	game := got[0]
	if game.seq != 2 || game.ID != "chg_2" || game.Type != "item.updated" || game.ItemType != "games" {
		t.Fatalf("game event = %+v", game)
	}

	if len(game.Fields) != 2 || game.Fields[0] != "ratingCount" || game.Fields[1] != "score" {
		t.Fatalf("game fields = %v", game.Fields)
	}

	movie := got[1]
	if movie.Type != "item.deleted" || movie.ItemType != "movies" || movie.MergedInto != "movie-2" ||
		movie.Fields != nil {
		t.Fatalf("movie event = %+v", movie)
	}

	if movie.ChangedAt != at.Add(2*time.Second).Format(time.RFC3339) {
		t.Fatalf("changedAt = %s", movie.ChangedAt)
	}
}

func TestChunksSplitAtBatchLimit(t *testing.T) {
	events := make([]batchEvent, batchLimit+1)
	parts := chunks(events, batchLimit)
	if len(parts) != 2 || len(parts[0]) != batchLimit || len(parts[1]) != 1 {
		t.Fatalf("chunks = %d parts (%d, %d)", len(parts), len(parts[0]), len(parts[1]))
	}

	if chunks(nil, batchLimit) != nil {
		t.Fatal("empty input produced a chunk")
	}
}
