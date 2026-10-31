//go:build integration

package ingest_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/flags"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

func newDB(t *testing.T) *database.DB {
	t.Helper()

	env := map[string]string{"DATABASE_URL": pgtest.Database(t, "")}
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

type envMap map[string]string

func (m envMap) Lookup(k string) (string, bool) { v, ok := m[k]; return v, ok }

func newFlags(t *testing.T, env map[string]string) *flags.Flags {
	t.Helper()

	ff, err := flags.New(flags.Config{Flags: run.All(), Env: envMap(env)})
	if err != nil {
		t.Fatal(err)
	}

	return ff
}

var allOn = map[string]string{"FLAG_INGEST_ENABLED": "true", "FLAG_INGEST_WIKIDATA": "true"}

// recorder stands in for *queue.Queue and keeps the successors dispatched in the chunk transaction.
type recorder struct{ jobs []queue.Job }

func (r *recorder) DispatchTx(_ context.Context, _ database.Executor, j queue.Job, _ ...queue.Option) error {
	r.jobs = append(r.jobs, j)

	return nil
}

func count(t *testing.T, db *database.DB, q string, args ...any) int {
	t.Helper()

	var n int
	if err := db.QueryRowContext(context.Background(), q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

// upsertStep stores targets Q1..Qtotal, perChunk per step, with the cursor {"next": n}.
func upsertStep(total, perChunk int, hook func()) run.Step {
	return func(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
		var cur struct{ Next int }
		if len(r.Cursor) > 0 {
			if err := json.Unmarshal(r.Cursor, &cur); err != nil {
				return run.Result{}, err
			}
		}

		if cur.Next == 0 {
			cur.Next = 1
		}

		res := run.Result{}

		for i := cur.Next; i < cur.Next+perChunk && i <= total; i++ {
			ok, err := store.Upsert(ctx, tx, r.ID, store.Record{
				Source: "wikidata", ExternalID: "Q" + itoa(i), Payload: json.RawMessage(`{"n":` + itoa(i) + `}`),
			})
			if err != nil {
				return res, err
			}

			res.Targets++
			res.Seen++

			if ok {
				res.Changed++
			}
		}

		next := cur.Next + perChunk
		res.Done = next > total
		res.Cursor, _ = json.Marshal(map[string]int{"next": next})

		if hook != nil {
			hook()
		}

		return res, nil
	}
}

func itoa(i int) string { b, _ := json.Marshal(i); return string(b) }

func newRunner(
	db *database.DB,
	d run.Dispatcher,
	ff *flags.Flags,
	lim run.Limits,
	steps map[string]run.Step,
) *run.Runner {
	return &run.Runner{DB: db, Queue: d, Flags: ff, Limits: lim, Steps: steps, Now: time.Now}
}
