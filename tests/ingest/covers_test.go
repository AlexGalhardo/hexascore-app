//go:build integration

package ingest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

const (
	coverA = "0192f1c6-0000-7000-8000-0000000000c1"
	coverB = "0192f1c6-0000-7000-8000-0000000000c2"
	coverC = "0192f1c6-0000-7000-8000-0000000000c3"
)

func TestCoversGC(t *testing.T) {
	db, disk := newDB(t), storage.NewMemory("")

	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	old, oldThumb, kept := "covers/a-1111111111111111.webp", "covers/a-1111111111111111-t.webp", "covers/b-2222.webp"
	for _, k := range []string{old, oldThumb, kept, "covers/a-new.webp"} {
		if err := storage.PutBytes(t.Context(), disk, k, []byte("webp")); err != nil {
			t.Fatal(err)
		}
	}

	exec(t, db, `INSERT INTO items (id, type, slug, wikidata_qid, cover_key, cover_thumb_key) VALUES
($1, 'game', 'a', 'Q1', $3, $4), ($2, 'game', 'b', 'Q2', $5, NULL)`, coverA, coverB, old, oldThumb, kept)
	// Replace A's cover, and release B's then point B at it again: only A's old pair may go.
	exec(t, db, `UPDATE items SET cover_key = 'covers/a-new.webp', cover_thumb_key = NULL WHERE id = $1`, coverA)
	exec(t, db, `UPDATE items SET cover_key = NULL WHERE id = $1`, coverB)
	exec(t, db, `UPDATE items SET cover_key = $2 WHERE id = $1`, coverB, kept)

	if n := count(t, db, `SELECT count(*) FROM cover_gc`); n != 3 {
		t.Fatalf("released keys = %d, want 3", n)
	}

	now := time.Now()

	if n, err := covers.GC(t.Context(), db, disk, now.Add(29*24*time.Hour)); err != nil || n != 0 {
		t.Fatalf("day 29: deleted %d, %v; want 0", n, err)
	}

	if n, err := covers.GC(t.Context(), db, disk, now.Add(31*24*time.Hour)); err != nil || n != 2 {
		t.Fatalf("day 31: deleted %d, %v; want 2", n, err)
	}

	for k, want := range map[string]bool{old: false, oldThumb: false, kept: true, "covers/a-new.webp": true} {
		if ok, _ := disk.Exists(t.Context(), k); ok != want {
			t.Errorf("%s exists = %v, want %v", k, ok, want)
		}
	}

	if n := count(t, db, `SELECT count(*) FROM cover_gc`); n != 0 {
		t.Errorf("cover_gc rows left = %d, want 0", n)
	}
}

type coverHosts struct{ jpeg []byte }

func (h coverHosts) RoundTrip(r *http.Request) (*http.Response, error) {
	status, body := http.StatusOK, h.jpeg
	if r.URL.Path == "/404.jpg" {
		status, body = http.StatusNotFound, nil
	}

	return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(body)), Header: http.Header{}}, nil
}

type tagEncoder struct{}

func (tagEncoder) Encode(_ context.Context, src []byte, v covers.Variant) ([]byte, error) {
	if _, _, err := image.DecodeConfig(bytes.NewReader(src)); err != nil {
		return nil, covers.ErrUnsupported
	}

	return []byte("RIFF\x00\x00\x00\x00WEBP" + string(rune('0'+v.Quality%10))), nil
}

func TestCoversStepStoresReportsAndDropsLooseCommons(t *testing.T) {
	db, disk := newDB(t), storage.NewMemory("")

	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	img := image.NewNRGBA(image.Rect(0, 0, 40, 60))
	img.Set(1, 1, color.White)

	var src bytes.Buffer
	if err := jpeg.Encode(&src, img, nil); err != nil {
		t.Fatal(err)
	}

	exec(t, db, `INSERT INTO items (id, type, slug, wikidata_qid, cover_key, cover_thumb_key, cover_source) VALUES
($1, 'game', 'a', 'Q1', NULL, NULL, NULL),
($2, 'movie', 'b', 'Q2', 'covers/b-0000000000000000.webp', 'covers/b-0000000000000000-t.webp', 'commons'),
($3, 'game', 'c', 'Q3', NULL, NULL, NULL)`, coverA, coverB, coverC)
	exec(t, db, `INSERT INTO source_records (source, external_id, payload, hash) VALUES
('steam', 'Q1', '{"header_image":"https://shared.akamai.steamstatic.com/a.jpg"}', repeat('1', 64)),
('wikidata', 'Q2', '{"claims":{"P18":[{"value":"Photo.jpg"}]}}', repeat('2', 64)),
('steam', 'Q3', '{"header_image":"https://shared.akamai.steamstatic.com/404.jpg"}', repeat('3', 64))`)

	ff := newFlags(t, map[string]string{"FLAG_INGEST_ENABLED": "true", "FLAG_INGEST_COVERS": "true"})
	job := &covers.Job{Disk: disk, HTTP: &http.Client{Transport: coverHosts{src.Bytes()}}, Encoder: tagEncoder{}}
	steps := map[string]run.Step{covers.JobName: job.Step}

	runOnce := func(params string) {
		t.Helper()

		r, err := run.Create(t.Context(), db, &recorder{}, run.Spec{
			Job: covers.JobName, Source: covers.JobName, Mode: run.ModeFull, Params: json.RawMessage(params),
		})
		if err != nil {
			t.Fatal(err)
		}

		if _, err := newRunner(db, &recorder{}, ff, run.Limits{}, steps).Execute(t.Context(), r.ID); err != nil {
			t.Fatal(err)
		}
	}

	// A run restricted to qids never strips covers.
	runOnce(`{"qids":["Q1","Q2"]}`)

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1 AND cover_source = 'commons'`, coverB); n != 1 {
		t.Fatal("a qids run dropped the loose Commons cover")
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1 AND cover_source = 'steam'
AND cover_key LIKE 'covers/' || $1 || '-%.webp' AND cover_thumb_key LIKE '%-t.webp'`, coverA); n != 1 {
		t.Fatal("item A has no steam cover")
	}

	runOnce(`{}`)

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1 AND cover_key IS NULL AND cover_source IS NULL`,
		coverB); n != 1 {
		t.Error("the full run kept the loose Commons cover")
	}

	if n := count(t, db, `SELECT count(*) FROM cover_gc WHERE key LIKE 'covers/b-%'`); n != 2 {
		t.Errorf("released B keys = %d, want 2", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM ingest_failures WHERE source = 'covers' AND external_id = 'Q3'`,
	); n != 1 {
		t.Error("the 404 cover of C is not in the dead letter")
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs WHERE job = 'covers' AND status = 'partial'`); n != 1 {
		t.Errorf("partial covers runs = %d, want 1 (the second)", n)
	}
}

func exec(t *testing.T, db *database.DB, q string, args ...any) {
	t.Helper()

	if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
		t.Fatal(err)
	}
}

func TestCoversStepRollbackLeavesReservedKeys(t *testing.T) {
	db, disk := newDB(t), storage.NewMemory("")

	img := image.NewNRGBA(image.Rect(0, 0, 40, 60))
	img.Set(1, 1, color.White)

	var src bytes.Buffer
	if err := jpeg.Encode(&src, img, nil); err != nil {
		t.Fatal(err)
	}

	exec(t, db, `INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', 'a', 'Q1')`, coverA)
	exec(t, db, `INSERT INTO source_records (source, external_id, payload, hash) VALUES
('steam', 'Q1', '{"header_image":"https://shared.akamai.steamstatic.com/a.jpg"}', repeat('1', 64))`)

	job := &covers.Job{
		Disk: disk, HTTP: &http.Client{Transport: coverHosts{src.Bytes()}}, Encoder: tagEncoder{},
		Reserve: covers.Reserver(db),
	}

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		if _, err := job.Step(t.Context(), tx, run.Run{Mode: run.ModeFull}); err != nil {
			return err
		}

		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("want the forced rollback")
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1 AND cover_key IS NOT NULL`, coverA); n != 0 {
		t.Fatal("the item update must have rolled back")
	}

	if n := count(t, db, `SELECT count(*) FROM cover_gc WHERE key LIKE 'covers/' || $1 || '-%'`, coverA); n != 2 {
		t.Errorf("reserved keys = %d, want 2 so the GC reaps the orphaned objects", n)
	}
}
