//go:build integration

package console_test

import (
	"context"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"

	"github.com/AlexGalhardo/hexascore-app/app/console"
)

func openDB(t *testing.T) *database.DB {
	t.Helper()

	db, err := database.Open(t.Context(), pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func TestMigrateLockedSerialises(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	entered, release, second := make(chan struct{}), make(chan struct{}), make(chan struct{})

	go func() {
		_ = console.WithMigrationLock(context.Background(), db, func(context.Context) error {
			close(entered)
			<-release

			return nil
		})
	}()

	<-entered

	done := make(chan error, 1)

	go func() {
		done <- console.WithMigrationLock(context.Background(), db, func(context.Context) error {
			close(second)
			return nil
		})
	}()

	select {
	case <-second:
		t.Fatal("second runner entered while the first held the lock")
	case <-time.After(500 * time.Millisecond):
	}

	close(release)

	select {
	case <-second:
	case <-time.After(5 * time.Second):
		t.Fatal("second runner never got the lock after release")
	}

	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestMigrationLockReleasedOnError(t *testing.T) {
	t.Parallel()

	db := openDB(t)
	ctx := t.Context()

	if err := console.WithMigrationLock(ctx, db, func(context.Context) error { return context.Canceled }); err == nil {
		t.Fatal("error from fn was swallowed")
	}

	ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := console.WithMigrationLock(ctx2, db, func(context.Context) error { return nil }); err != nil {
		t.Fatalf("lock not released after a failing run: %v", err)
	}
}
