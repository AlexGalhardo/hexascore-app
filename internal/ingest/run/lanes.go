package run

import (
	"encoding/json"
	"slices"
	"sync/atomic"

	"github.com/AlexGalhardo/vukem/queue"
)

// Queue is the queue every ingestion run uses unless lanes are on.
const Queue = "ingest"

// Categories are the item types that get a lane of their own: a queue per category, so the worker runs one job of
// each in parallel instead of letting one long run hold every slot (owner 2026-11-04).
var Categories = []string{"movie", "series", "game", "book"}

var lanes atomic.Bool

// SetLanes switches the per-category queues on or off (env INGEST_LANES, read at boot). Off, every run goes to
// Queue; on, a run whose params name exactly one category goes to LaneQueue of it. The worker must consume those
// queues before this is switched on (docs/runbooks/worker.md).
func SetLanes(on bool) { lanes.Store(on) }

// LaneQueue is the queue of one category.
func LaneQueue(category string) string { return Queue + "-" + category }

// Types reads the "types" key of a run's params; nil when absent or unreadable.
func Types(params json.RawMessage) []string {
	var p struct {
		Types []string `json:"types"`
	}

	if len(params) == 0 || json.Unmarshal(params, &p) != nil {
		return nil
	}

	return p.Types
}

// QueueFor is the queue a run with these params is dispatched to.
func QueueFor(params json.RawMessage) string {
	if !lanes.Load() {
		return Queue
	}

	if t := Types(params); len(t) == 1 && slices.Contains(Categories, t[0]) {
		return LaneQueue(t[0])
	}

	return Queue
}

func onQueue(params json.RawMessage) queue.Option { return queue.OnQueue(QueueFor(params)) }
