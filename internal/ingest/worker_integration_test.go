//go:build integration

package ingest_test

import (
	"context"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// RFC-001 section 8: the worker writes a heartbeat row every minute; the web's readiness check reads it.
func TestWorkerHeartbeat(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)

	if err := ingest.CheckWorker(ctx, db, now); err == nil {
		t.Fatal("no heartbeat yet: the check must fail, the worker never ran")
	}

	if err := ingest.WriteWorkerHeartbeat(ctx, db); err != nil {
		t.Fatal(err)
	}

	at, ok, err := ingest.ReadWorkerHeartbeat(ctx, db)
	if err != nil || !ok {
		t.Fatalf("read = %v, %v, %v", at, ok, err)
	}

	if err := ingest.CheckWorker(ctx, db, at.Add(ingest.WorkerStaleAfter-time.Second)); err != nil {
		t.Fatalf("a fresh heartbeat must pass: %v", err)
	}

	if err := ingest.CheckWorker(ctx, db, at.Add(ingest.WorkerStaleAfter)); err == nil {
		t.Fatal("a heartbeat three minutes old must fail")
	}

	// the heartbeat does not disturb the pressure signal or a host row
	if _, ok, err := ingest.ReadSignal(ctx, db); err != nil || ok {
		t.Fatalf("signal = %v, %v: the worker heartbeat must not be read as one", ok, err)
	}
}
