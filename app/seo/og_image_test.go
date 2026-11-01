package seo_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/seo"
)

var red = color.RGBA{R: 0xff, A: 0xff}

func solid(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}

	return img
}

func pngBytes(t *testing.T, img image.Image) []byte {
	t.Helper()

	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}

	return b.Bytes()
}

func same(a, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()

	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func TestCard_CoverOnPanel(t *testing.T) {
	t.Parallel()

	card := seo.Card(solid(400, 600, red))
	if b := card.Bounds(); b.Dx() != seo.CardWidth || b.Dy() != seo.CardHeight {
		t.Fatalf("size = %v, want %dx%d", b, seo.CardWidth, seo.CardHeight)
	}

	if got := card.At(seo.CardWidth-10, seo.CardHeight/2); !same(got, seo.Panel) {
		t.Errorf("panel pixel = %v, want %v", got, seo.Panel)
	}

	// The cover keeps its own size (never upscaled) and sits with equal top, bottom and left margins.
	if got := card.At(15+200, 315); !same(got, red) {
		t.Errorf("cover pixel = %v, want red", got)
	}
	if got := card.At(5, 315); !same(got, seo.Panel) {
		t.Errorf("left margin pixel = %v, want the panel", got)
	}
}

func TestCard_NoCoverIsPanelOnly(t *testing.T) {
	t.Parallel()

	card := seo.Card(nil)
	for _, p := range []image.Point{{0, 0}, {600, 315}, {1199, 629}} {
		if got := card.At(p.X, p.Y); !same(got, seo.Panel) {
			t.Errorf("pixel %v = %v, want the panel", p, got)
		}
	}
}

func TestCards_RendersOnceAndCachesOnTheDisk(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	disk := storage.NewMemory("")
	if err := disk.Put(ctx, "covers/a.png", bytes.NewReader(pngBytes(t, solid(40, 60, red)))); err != nil {
		t.Fatal(err)
	}

	cards := seo.NewCards(disk)
	first, err := cards.PNG(ctx, "items", "id1", "covers/a.png")
	if err != nil {
		t.Fatal(err)
	}

	img, err := png.Decode(bytes.NewReader(first))
	if err != nil {
		t.Fatal(err)
	}
	if got := img.At(seo.CardHeight/2-30+20, seo.CardHeight/2); !same(got, red) {
		t.Errorf("cover pixel = %v, want red", got)
	}

	keys, err := disk.List(ctx, "og/items")
	if err != nil || len(keys) != 1 || !strings.HasPrefix(keys[0], "og/items/id1-") {
		t.Fatalf("cached keys = %v (%v), want one og/items/id1-<hash>.png", keys, err)
	}

	// The second call reads the cached card, so a deleted cover does not change it.
	if err := disk.Delete(ctx, "covers/a.png"); err != nil {
		t.Fatal(err)
	}
	second, err := cards.PNG(ctx, "items", "id1", "covers/a.png")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Error("second render differs from the cached card")
	}

	// A new cover key is a new cache entry.
	if _, err := cards.PNG(ctx, "items", "id1", ""); err != nil {
		t.Fatal(err)
	}
	if keys, _ := disk.List(ctx, "og/items"); len(keys) != 2 {
		t.Errorf("cached keys = %v, want two", keys)
	}
}

func TestCards_UndecodableCoverFallsBackToThePanel(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	disk := storage.NewMemory("")
	if err := disk.Put(ctx, "covers/demo.svg", strings.NewReader("<svg/>")); err != nil {
		t.Fatal(err)
	}

	for _, d := range []storage.Disk{disk, nil} {
		out, err := seo.NewCards(d).PNG(ctx, "lists", "l1", "covers/demo.svg")
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(out))
		if err != nil {
			t.Fatal(err)
		}
		if got := img.At(100, 315); !same(got, seo.Panel) {
			t.Errorf("disk %v: pixel = %v, want the panel", d, got)
		}
	}
}
