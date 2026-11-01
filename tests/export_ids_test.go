//go:build integration

package tests

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestExportIDs(t *testing.T) {
	now := time.Date(2026, 11, 2, 7, 0, 0, 0, time.UTC)
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:   map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
		Clock: vukemtest.NewFakeClock(now).Now,
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	pub, hidden := uuid.NewV7().String(), uuid.NewV7().String()

	if _, err := db.SQL().ExecContext(ctx, `
INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior, updated_at, visibility) VALUES
  ($1, 'game', 'zelda', 'Q1', 0.5, '2026-11-01T10:00:00Z', 'public'),
  ($2, 'game', 'secret', 'Q2', 0.1, '2026-11-01T10:00:00Z', 'takedown')`, pub, hidden); err != nil {
		t.Fatal(err)
	}

	disk, err := storage.NewLocal(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = disk.Close() })

	for _, k := range []string{"exports/games_ids_2026_08_03.json.gz", "exports/games_ids_2026_08_04.json.gz"} {
		if err := disk.Put(ctx, k, strings.NewReader("old")); err != nil {
			t.Fatal(err)
		}
	}

	if err := jobs.RunExportIDs(ctx, db.SQL(), disk, now); err != nil {
		t.Fatal(err)
	}

	r, err := disk.Open(ctx, "exports/games_ids_2026_11_02.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	zr, err := gzip.NewReader(r)
	if err != nil {
		t.Fatal(err)
	}

	sc := bufio.NewScanner(zr)

	var lines []map[string]any

	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatal(err)
		}

		lines = append(lines, m)
	}

	if len(lines) != 1 {
		t.Fatalf("%d lines, want 1 (takedown excluded): %v", len(lines), lines)
	}

	want := map[string]any{
		"id":        pub,
		"slug":      "zelda",
		"wikidata":  "Q1",
		"updatedAt": "2026-11-01T10:00:00Z",
	}
	for k, v := range want {
		if lines[0][k] != v {
			t.Errorf("%s = %v, want %v", k, lines[0][k], v)
		}
	}

	// popularity_prior stays out of the served files until the owner answers Q-A11 (A-07).
	if _, ok := lines[0]["popularity"]; ok {
		t.Errorf("export line carries popularity: %v", lines[0])
	}

	for k, want := range map[string]bool{
		"exports/games_ids_2026_08_03.json.gz":  false, // day -91
		"exports/games_ids_2026_08_04.json.gz":  true,  // day -90
		"exports/movies_ids_2026_11_02.json.gz": true,  // an empty type still gets a file
	} {
		got, err := disk.Exists(ctx, k)
		if err != nil || got != want {
			t.Errorf("exists(%s) = %v, %v; want %v", k, got, err, want)
		}
	}
}
