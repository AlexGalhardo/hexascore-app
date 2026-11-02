//go:build integration

package tests

import (
	"context"
	"io"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

type putCounter struct {
	*storage.Memory

	mu   sync.Mutex
	puts []string
}

func (d *putCounter) Put(ctx context.Context, p string, r io.Reader) error {
	d.mu.Lock()
	d.puts = append(d.puts, p)
	d.mu.Unlock()

	return d.Memory.Put(ctx, p, r)
}

func TestSourceRecordHistory_PartitionsAndColdExport(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

	// The migration pre-created the current month and three more; make the older ones the job's own work.
	if err := store.EnsureHistoryPartitions(ctx, db, now.AddDate(0, -4, 0), 4); err != nil {
		t.Fatal(err)
	}

	for i, at := range []string{"2026-07-10T10:00:00Z", "2026-07-11T10:00:00Z", "2026-09-05T10:00:00Z", "2026-10-31T23:59:59Z"} {
		if _, err := db.ExecContext(ctx, `INSERT INTO source_record_history
(source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run)
VALUES ('wikidata', $1, 'h', '{"v":1}', $2, $2, 1)`, "Q"+string(rune('1'+i)), at); err != nil {
			t.Fatal(err)
		}
	}

	var where string

	for q, want := range map[string]string{"Q1": "source_record_history_2026_07", "Q3": "source_record_history_2026_09", "Q4": "source_record_history_2026_10"} {
		if err := db.QueryRowContext(ctx, `SELECT tableoid::regclass::text FROM source_record_history WHERE external_id = $1`, q).
			Scan(&where); err != nil ||
			where != want {
			t.Errorf("%s landed in %q (%v), want %s", q, where, err, want)
		}
	}

	disk := &putCounter{Memory: storage.NewMemory("")}

	res, err := jobs.RunExportSourceHistory(ctx, db, disk, now)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{"2026-07", "2026-09", "2026-10"}; !slices.Equal(res.Exported, want) {
		t.Errorf("exported %v, want %v", res.Exported, want)
	}

	// 30 days before Nov 20 is Oct 21: July and September are fully older, October is not.
	if want := []string{"2026-06", "2026-07", "2026-08", "2026-09"}; !slices.Equal(res.Dropped, want) {
		t.Errorf("dropped %v, want %v", res.Dropped, want)
	}

	for _, m := range []string{"2026-07", "2026-09", "2026-10"} {
		if ok, _ := disk.Exists(ctx, "exports/source_record_history/"+m+".ndjson.gz"); !ok {
			t.Errorf("object for %s missing", m)
		}
	}

	puts := len(disk.puts)

	again, err := jobs.RunExportSourceHistory(ctx, db, disk, now)
	if err != nil || len(again.Exported) != 0 || len(again.Dropped) != 0 || len(disk.puts) != puts {
		t.Errorf("second run: %+v, %v, puts %d -> %d, want a no-op", again, err, puts, len(disk.puts))
	}

	// The October rows are still queryable (retention window) and a live upsert lands in the current month.
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM source_record_history`).Scan(&n); err != nil || n != 1 {
		t.Errorf("rows left = %d (%v), want 1", n, err)
	}
}
