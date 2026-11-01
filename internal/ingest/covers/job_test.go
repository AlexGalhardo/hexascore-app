package covers_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

const sourceURL = "https://shared.akamai.steamstatic.com/header.jpg"

// stubEncoder stands in for cwebp: it sniffs the source like the real encoder and returns a tagged WebP shell.
type stubEncoder struct{}

func (stubEncoder) Encode(_ context.Context, src []byte, v covers.Variant) ([]byte, error) {
	if _, _, err := image.DecodeConfig(bytes.NewReader(src)); err != nil {
		return nil, covers.ErrUnsupported
	}

	size := []byte{byte(v.MaxW), byte(v.MaxH)}

	return append(append([]byte("RIFF\x00\x00\x00\x00WEBP"), size...), bytes.Repeat([]byte{1}, v.MaxW)...), nil
}

// countingDisk records the paths written.
type countingDisk struct {
	*storage.Memory

	mu   sync.Mutex
	puts []string
}

func (d *countingDisk) Put(ctx context.Context, p string, r io.Reader) error {
	d.mu.Lock()
	d.puts = append(d.puts, p)
	d.mu.Unlock()

	return d.Memory.Put(ctx, p, r)
}

func newDisk() *countingDisk { return &countingDisk{Memory: storage.NewMemory("")} }

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// sourceBytes is the image every fake host serves.
var sourceBytes = func() []byte {
	var b bytes.Buffer
	if err := jpeg.Encode(&b, gradient(60, 90, 255), nil); err != nil {
		panic(err)
	}

	return b.Bytes()
}()

func reply(status int, body []byte) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(body)), Header: http.Header{}}
}

// hosts serves sourceBytes for every GET and counts them; override sets other answers by path.
type hosts struct {
	gets     atomic.Int32
	override map[string]func() *http.Response
}

func (h *hosts) client() *http.Client {
	return &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		h.gets.Add(1)

		if f, ok := h.override[r.URL.Path]; ok {
			return f(), nil
		}

		return reply(http.StatusOK, sourceBytes), nil
	})}
}

func newJob(h *hosts, disk storage.Disk) *covers.Job {
	return &covers.Job{Disk: disk, HTTP: h.client(), Encoder: stubEncoder{}}
}

func key(id string, suffix string) string {
	sum := sha256.Sum256(sourceBytes)

	return "covers/" + id + "-" + hex.EncodeToString(sum[:])[:16] + suffix + ".webp"
}

// zelda is a Store target; covers live under the slug keys covers/games/zelda[-t].webp (item 18).
func zelda(cur, curThumb string) covers.Target {
	return covers.Target{ItemID: "item-1", Type: "game", Slug: "zelda", Key: cur, ThumbKey: curThumb}
}

func steamRecords(qid, header string) []store.Record {
	return []store.Record{{Source: "steam", ExternalID: qid, Payload: []byte(`{"header_image":"` + header + `"}`)}}
}

func TestStoreCoverUploadsFullAndThumbUnderSlugKeys(t *testing.T) {
	h, disk := &hosts{}, newDisk()

	full, thumb, err := newJob(h, disk).Store(context.Background(), zelda("", ""), sourceURL, false)
	if err != nil {
		t.Fatal(err)
	}

	if full != "covers/games/zelda.webp" || thumb != "covers/games/zelda-t.webp" {
		t.Fatalf("keys = %q, %q", full, thumb)
	}

	if !slices.Equal(disk.puts, []string{full, thumb}) || h.gets.Load() != 1 {
		t.Fatalf("puts = %v, gets = %d", disk.puts, h.gets.Load())
	}

	big, _ := storage.Get(context.Background(), disk, full)
	small, _ := storage.Get(context.Background(), disk, thumb)

	if !bytes.HasPrefix(big, []byte("RIFF")) || string(big[8:12]) != "WEBP" || len(small) >= len(big)/3 {
		t.Fatalf("full %d bytes, thumb %d bytes", len(big), len(small))
	}
}

func TestStoreCoverSkipsUploadWhenSameSourceIsStored(t *testing.T) {
	h, disk := &hosts{}, newDisk()

	full, thumb, err := newJob(h, disk).
		Store(context.Background(), zelda(key("item-1", ""), key("item-1", "-t")), sourceURL, true)
	if err != nil || full != key("item-1", "") || thumb != key("item-1", "-t") {
		t.Fatalf("got %q %q %v", full, thumb, err)
	}

	if len(disk.puts) != 0 || h.gets.Load() != 1 {
		t.Fatalf("puts = %v, gets = %d", disk.puts, h.gets.Load())
	}
}

func TestStoreCoverAddsOnlyMissingThumb(t *testing.T) {
	h, disk := &hosts{}, newDisk()

	full, thumb, err := newJob(h, disk).Store(context.Background(), zelda(key("item-1", ""), ""), sourceURL, true)
	if err != nil || full != key("item-1", "") || thumb != "covers/games/zelda-t.webp" {
		t.Fatalf("got %q %q %v", full, thumb, err)
	}

	if !slices.Equal(disk.puts, []string{"covers/games/zelda-t.webp"}) {
		t.Fatalf("puts = %v", disk.puts)
	}
}

func TestStoreCoverReuploadsWhenSourceChanged(t *testing.T) {
	h, disk := &hosts{}, newDisk()

	full, thumb, err := newJob(h, disk).Store(context.Background(),
		zelda("covers/item-1-0000000000000000.webp", "covers/item-1-0000000000000000-t.webp"), sourceURL, false)
	if err != nil || full != "covers/games/zelda.webp" || thumb != "covers/games/zelda-t.webp" || len(disk.puts) != 2 {
		t.Fatalf("got %q %q %v, puts %v", full, thumb, err, disk.puts)
	}
}

// The thumbnail slug key of "zelda" is the full key of an item slugged "zelda-t": that item keeps it.
func TestStoreCoverThumbKeepsAHashedKeyWhenItsSlugKeyIsTaken(t *testing.T) {
	h, disk := &hosts{}, newDisk()
	target := zelda("", "")
	target.ThumbTaken = true

	full, thumb, err := newJob(h, disk).Store(context.Background(), target, sourceURL, false)
	if err != nil || full != "covers/games/zelda.webp" || thumb != key("item-1", "-t") {
		t.Fatalf("got %q %q %v", full, thumb, err)
	}
}

func TestStoreCoverRejectsAFailedDownload(t *testing.T) {
	h := &hosts{override: map[string]func() *http.Response{
		"/404.jpg": func() *http.Response { return reply(http.StatusNotFound, nil) },
	}}

	_, _, err := newJob(h, newDisk()).Store(context.Background(), zelda("", ""), "https://cdn.example/404.jpg", false)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("err = %v, want a 404 error", err)
	}
}

func TestCommonsCoverOnlyFromFilmPoster(t *testing.T) {
	p18 := []store.Record{{Source: "wikidata", ExternalID: "Q1", Payload: []byte(
		`{"claims":{"P18":[{"value":"Mario arcade.jpg","rank":"normal"}]}}`)}}
	if f := covers.CommonsFile("Q1", p18); f != "" {
		t.Fatalf("P18 gave %q, want no Commons file", f)
	}

	both := []store.Record{{Source: "wikidata", ExternalID: "Q1", Payload: []byte(
		`{"claims":{"P18":[{"value":"Mario arcade.jpg"}],"P3383":[{"value":"Poster.jpg"}]}}`)}}
	if f := covers.CommonsFile("Q1", both); f != "Poster.jpg" {
		t.Fatalf("P3383 gave %q, want Poster.jpg", f)
	}
}

func TestCommonsCoverIsLookedUpWhenNoSourceNamesOne(t *testing.T) {
	h := &hosts{override: map[string]func() *http.Response{
		"/w/api.php": func() *http.Response {
			return reply(
				http.StatusOK,
				[]byte(`{"query":{"pages":[{"imageinfo":[{"url":"https://upload.wikimedia.org/a.jpg",`+
					`"thumburl":"https://upload.wikimedia.org/600px-a.jpg","descriptionurl":"https://commons.wikimedia.org/wiki/File:a.jpg",`+
					`"extmetadata":{"Artist":{"value":"<b>Ann</b>"},"LicenseShortName":{"value":"CC BY 4.0"}}}]}]}}`),
			)
		},
	}}
	recs := []store.Record{
		{Source: "wikidata", ExternalID: "Q2", Payload: []byte(`{"claims":{"P3383":[{"value":"a.jpg"}]}}`)},
	}

	p, err := newJob(
		h,
		newDisk(),
	).Cover(context.Background(), covers.Target{ItemID: "i2", QID: "Q2", Type: "movie", Records: recs})
	if err != nil || p == nil {
		t.Fatalf("patch %v, err %v", p, err)
	}

	if p.Source != "commons" || p.Attribution["author"] != "Ann" || p.Attribution["license"] != "CC BY 4.0" {
		t.Fatalf("patch = %+v", p)
	}
}

func TestRunIsolatesPerItemFailures(t *testing.T) {
	h := &hosts{override: map[string]func() *http.Response{
		"/broken.jpg": func() *http.Response { return reply(http.StatusOK, []byte("not an image")) },
		"/404.jpg":    func() *http.Response { return reply(http.StatusNotFound, nil) },
	}}
	targets := []covers.Target{
		{ItemID: "broken", QID: "Q1", Type: "game", Records: steamRecords("Q1", "https://cdn.example/broken.jpg")},
		{ItemID: "gone", QID: "Q2", Type: "game", Records: steamRecords("Q2", "https://cdn.example/404.jpg")},
		{ItemID: "ok", QID: "Q3", Type: "game", Slug: "ok", Records: steamRecords("Q3", sourceURL)},
	}

	out := newJob(h, newDisk()).Run(context.Background(), targets)

	if out[0].Err == nil || !errors.Is(out[0].Err, covers.ErrUnsupported) || out[1].Err == nil {
		t.Fatalf("errors = %v, %v", out[0].Err, out[1].Err)
	}

	ok := out[2]
	if ok.Err != nil || ok.Patch == nil || ok.Patch.Key != "covers/games/ok.webp" ||
		ok.Patch.ThumbKey != "covers/games/ok-t.webp" ||
		ok.Patch.Source != "steam" {
		t.Fatalf("ok = %+v", ok)
	}
}

func TestRunBackfillsThumbForLegacyCover(t *testing.T) {
	h, disk := &hosts{}, newDisk()
	target := covers.Target{
		ItemID: "item-1", QID: "Q1", Type: "game", Slug: "zelda", Records: steamRecords("Q1", sourceURL),
		Key: key("item-1", ""), Source: "steam",
	}

	out := newJob(h, disk).Run(context.Background(), []covers.Target{target})
	if p := out[0].Patch; out[0].Err != nil || p == nil || p.Key != key("item-1", "") ||
		p.ThumbKey != "covers/games/zelda-t.webp" {
		t.Fatalf("out = %+v", out[0])
	}

	if !slices.Equal(disk.puts, []string{"covers/games/zelda-t.webp"}) {
		t.Fatalf("puts = %v", disk.puts)
	}
}

func TestRunSkipsItemWithBothSizesCurrent(t *testing.T) {
	h := &hosts{}
	target := covers.Target{
		ItemID: "item-1", QID: "Q1", Type: "game", Records: steamRecords("Q1", sourceURL),
		Key: key("item-1", ""), ThumbKey: key("item-1", "-t"), Source: "steam",
	}

	out := newJob(h, newDisk()).Run(context.Background(), []covers.Target{target})
	if out[0].Err != nil || out[0].Patch != nil || h.gets.Load() != 0 {
		t.Fatalf("out = %+v, gets = %d", out[0], h.gets.Load())
	}
}

func TestRunReplacesALesserSource(t *testing.T) {
	h := &hosts{}
	recs := append(steamRecords("Q1", sourceURL), store.Record{
		Source:     "igdb",
		ExternalID: "Q1",
		Payload: []byte(
			`{"cover":"https://images.igdb.com/igdb/image/upload/t_cover_big_2x/x.jpg","url":"https://www.igdb.com/games/x"}`,
		),
	})
	target := covers.Target{
		ItemID: "item-1", QID: "Q1", Type: "game", Records: recs,
		Key: key("item-1", ""), ThumbKey: key("item-1", "-t"), Source: "steam",
	}

	out := newJob(h, newDisk()).Run(context.Background(), []covers.Target{target})
	if p := out[0].Patch; out[0].Err != nil || p == nil || p.Source != "igdb" || p.Attribution["source"] != "IGDB" {
		t.Fatalf("out = %+v", out[0])
	}
}

func TestRunProcessesItemsConcurrently(t *testing.T) {
	var (
		now, peak atomic.Int32
		once      sync.Once
		overlap   = make(chan struct{})
	)

	// Every download waits until two overlap (or 5 s pass), so a serial run shows a peak of 1.
	client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		n := now.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}

		if n >= 2 {
			once.Do(func() { close(overlap) })
		}

		select {
		case <-overlap:
		case <-time.After(5 * time.Second):
		}

		now.Add(-1)

		return reply(http.StatusOK, sourceBytes), nil
	})}

	targets := make([]covers.Target, 20)
	for i := range targets {
		id := string(rune('a' + i))
		targets[i] = covers.Target{ItemID: id, QID: "Q1", Type: "game", Records: steamRecords("Q1", sourceURL)}
	}

	j := &covers.Job{Disk: newDisk(), HTTP: client, Encoder: stubEncoder{}, Workers: 4}

	out := j.Run(context.Background(), targets)
	for _, o := range out {
		if o.Err != nil || o.Patch == nil {
			t.Fatalf("out = %+v", o)
		}
	}

	if peak.Load() < 2 || peak.Load() > 4 {
		t.Fatalf("peak concurrency %d, want 2..4", peak.Load())
	}
}

func TestStoreReservesEachKeyBeforeItsPut(t *testing.T) {
	h, disk := &hosts{}, newDisk()
	j := newJob(h, disk)

	var order []string

	j.Reserve = func(_ context.Context, k string) error {
		order = append(order, "reserve "+k)

		return nil
	}

	full, thumb, err := j.Store(context.Background(), zelda("", ""), sourceURL, false)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{"reserve " + full, "reserve " + thumb}; !slices.Equal(order, want) {
		t.Fatalf("reserved = %v, want %v", order, want)
	}

	if len(disk.puts) != 2 {
		t.Fatalf("puts = %v", disk.puts)
	}
}

func TestStoreWritesNothingWhenReserveFails(t *testing.T) {
	h, disk := &hosts{}, newDisk()
	j := newJob(h, disk)
	j.Reserve = func(context.Context, string) error { return errors.New("db down") }

	if _, _, err := j.Store(context.Background(), zelda("", ""), sourceURL, false); err == nil {
		t.Fatal("want the reserve error")
	}

	if len(disk.puts) != 0 {
		t.Fatalf("puts = %v, want none", disk.puts)
	}
}
