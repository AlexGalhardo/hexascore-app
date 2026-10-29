//go:build integration

package backup_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/crypt"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/backup"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestCopyMonthlyAndPrune(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	enc, err := crypt.New(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}

	day1 := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	disk := storage.NewMemory("")
	r := &backup.Runner{DB: db, Disk: disk, Enc: enc, Env: "test", Now: func() time.Time { return day1 }}
	ctx := t.Context()

	plain := make([]byte, 9<<20) // spans three 4 MiB frames
	_, _ = rand.Read(plain)

	var sealed bytes.Buffer

	src := backup.Result{Key: backup.ObjectKey("daily", "test", day1)}
	fw := backup.NewFrameWriter(&sealed, enc, src.Key)

	if _, err := fw.Write(plain); err != nil {
		t.Fatal(err)
	}

	if err := fw.Close(); err != nil {
		t.Fatal(err)
	}

	if err := disk.Put(ctx, src.Key, &sealed); err != nil {
		t.Fatal(err)
	}

	m, err := r.CopyMonthly(ctx, src)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}

	obj, err := disk.Open(ctx, m.Key)
	if err != nil {
		t.Fatal(err)
	}

	got, err := io.ReadAll(backup.NewFrameReader(obj, enc, m.Key))
	_ = obj.Close()

	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("monthly copy does not decrypt to the daily dump (err %v)", err)
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM backup_runs WHERE id = $1 AND kind = 'monthly'`, m.ID).
		Scan(&status); err != nil || status != "uploaded" {
		t.Fatalf("monthly row status = %q (%v)", status, err)
	}

	for i := range 31 {
		if err := disk.Put(
			ctx,
			backup.ObjectKey("daily", "test", day1.AddDate(0, 0, -i-1)),
			bytes.NewReader(nil),
		); err != nil {
			t.Fatal(err)
		}
	}

	n, err := r.Prune(ctx)
	if err != nil || n != 2 {
		t.Fatalf("pruned %d (%v), want 2 of 32 daily", n, err)
	}

	if ok, _ := disk.Exists(ctx, src.Key); !ok {
		t.Fatal("prune deleted the newest daily dump")
	}
}
