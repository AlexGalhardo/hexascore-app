package covers_test

import (
	"context"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
)

type memHashes struct {
	mu   sync.Mutex
	rows map[string]covers.HashRow
}

func (m *memHashes) Same(_ context.Context, key string, dhash uint64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.rows[key]

	return ok && covers.Distance(r.DHash, dhash) <= covers.NearDuplicate, nil
}

func (m *memHashes) Put(_ context.Context, r covers.HashRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.rows == nil {
		m.rows = map[string]covers.HashRow{}
	}

	m.rows[r.Key] = r

	return nil
}

type memRights map[string]string

func (m memRights) Allowed(_ context.Context, source string) (string, bool, error) {
	l, ok := m[source]

	return l, ok && l != "pending", nil
}

// The same picture offered by a second source is stored once: no encode, no upload, the keys stay.
func TestSamePictureFromTwoSourcesIsStoredOnce(t *testing.T) {
	h, disk, hashes := &hosts{}, newDisk(), &memHashes{}
	j := newJob(h, disk)
	j.Hashes = hashes
	recs := steamRecords("Q1", "https://shared.akamai.steamstatic.com/a.jpg")

	first, err := j.Cover(context.Background(), covers.Target{
		ItemID: "item-1", QID: "Q1", Type: "game", Slug: "zelda", Records: recs,
	})
	if err != nil || first == nil || len(disk.puts) != 2 {
		t.Fatalf("first cover: %+v, %v, puts %v", first, err, disk.puts)
	}

	if row, ok := hashes.rows[first.Key]; !ok || row.Source != "steam" || row.Bytes == 0 {
		t.Fatalf("hash row = %+v, %v", row, ok)
	}

	// Another source stored this item's cover before; steam now offers the same picture (a rescaled copy).
	second, err := j.Cover(context.Background(), covers.Target{
		ItemID: "item-1", QID: "Q1", Type: "game", Slug: "zelda", Records: recs,
		Key: first.Key, ThumbKey: first.ThumbKey, Source: "tmdb",
	})
	if err != nil || second == nil {
		t.Fatalf("second cover: %+v, %v", second, err)
	}

	if len(disk.puts) != 2 || second.Key != first.Key || second.ThumbKey != first.ThumbKey {
		t.Errorf("same picture written again: puts %v, keys %q %q", disk.puts, second.Key, second.ThumbKey)
	}
}

// A source whose licence grants no image rights is skipped: nothing is downloaded or stored.
func TestSourceWithoutImageRightsIsSkipped(t *testing.T) {
	h, disk := &hosts{}, newDisk()
	j := newJob(h, disk)
	j.Rights = memRights{"steam": "pending"}

	p, err := j.Cover(context.Background(), covers.Target{
		ItemID: "item-1", QID: "Q1", Type: "game", Slug: "zelda",
		Records: steamRecords("Q1", "https://shared.akamai.steamstatic.com/a.jpg"),
	})
	if err != nil || p != nil || h.gets.Load() != 0 || len(disk.puts) != 0 {
		t.Fatalf("patch %+v, err %v, gets %d, puts %v", p, err, h.gets.Load(), disk.puts)
	}

	j.Rights = memRights{"steam": "STEAM-WEB-API-TERMS"}
	if p, err = j.Cover(context.Background(), covers.Target{
		ItemID: "item-1", QID: "Q1", Type: "game", Slug: "zelda",
		Records: steamRecords("Q1", "https://shared.akamai.steamstatic.com/a.jpg"),
	}); err != nil || p == nil {
		t.Fatalf("allowed source: %+v, %v", p, err)
	}
}

func TestDHashSurvivesRescaleAndSeparatesPictures(t *testing.T) {
	a := gradient(120, 180, 255)
	small := gradient(40, 60, 255)
	other := image.NewNRGBA(image.Rect(0, 0, 120, 180))

	for y := range 180 {
		for x := range 120 {
			other.SetNRGBA(
				x,
				y,
				color.NRGBA{uint8(255 - x*255/120), uint8(255 - x*255/120), uint8(255 - x*255/120), 255},
			)
		}
	}

	if d := covers.Distance(covers.DHash(a), covers.DHash(small)); d > covers.NearDuplicate {
		t.Errorf("rescaled copy at distance %d", d)
	}

	if d := covers.Distance(covers.DHash(a), covers.DHash(other)); d <= covers.NearDuplicate {
		t.Errorf("different picture at distance %d", d)
	}
}

// On Linux the child runs under ulimit -t / -v: the fake script reports the limits it sees, and a busy loop is
// killed by the CPU limit instead of running to the wall-clock timeout.
func TestCwebpRunsUnderResourceLimits(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("rlimits are applied on Linux only")
	}

	if _, err := exec.LookPath("/bin/sh"); err != nil {
		t.Skip("no /bin/sh")
	}

	dir := t.TempDir()
	report := filepath.Join(dir, "report.sh")
	spin := filepath.Join(dir, "spin.sh")

	if err := os.WriteFile(
		report,
		[]byte("#!/bin/sh\nfor a; do out=$a; done\necho \"t=$(ulimit -t) v=$(ulimit -v)\" > \"$out\"\n"),
		0o700,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(spin, []byte("#!/bin/sh\nwhile :; do :; done\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	src := pngBytes(t, gradient(10, 10, 255))

	out, err := (&covers.Cwebp{Path: report}).Encode(context.Background(), src, covers.Full)
	if err != nil || strings.TrimSpace(string(out)) != "t=30 v=524288" {
		t.Fatalf("limits = %q, %v, want t=30 v=524288", out, err)
	}

	start := time.Now()

	_, err = (&covers.Cwebp{Path: spin, CPUSeconds: 1, Timeout: 20 * time.Second}).
		Encode(context.Background(), src, covers.Full)
	if err == nil || time.Since(start) > 10*time.Second {
		t.Errorf("busy child: err %v after %v, want a kill by the CPU limit", err, time.Since(start))
	}
}
