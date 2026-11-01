package ingest_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
)

// pngHeader is a valid PNG signature and IHDR declaring w x h, and no pixel data: decoding it fails, reading its
// header does not, so an ErrTooLarge result proves the rejection came before any decode.
func pngHeader(t *testing.T, w, h uint32) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewGray(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}

	hdr := buf.Bytes()[:33]
	binary.BigEndian.PutUint32(hdr[16:], w)
	binary.BigEndian.PutUint32(hdr[20:], h)
	binary.BigEndian.PutUint32(hdr[29:], crc32.ChecksumIEEE(hdr[12:29]))

	return hdr
}

func TestSecurity_A05_ImageBomb(t *testing.T) {
	// The binary does not exist: a request that got past the size checks would fail with ErrNoEncoder instead.
	enc := &covers.Cwebp{Path: "cwebp-that-does-not-exist"}

	cases := []struct {
		name string
		src  []byte
	}{
		{"one pixel over 25 MP", pngHeader(t, 5000, 5001)},
		{"65535 x 65535 header", pngHeader(t, 65535, 65535)},
		{"one byte over 15 MB", make([]byte, covers.MaxSourceBytes+1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := enc.Encode(t.Context(), tc.src, covers.Full)
			if !errors.Is(err, covers.ErrTooLarge) {
				t.Fatalf("error = %v, want ErrTooLarge", err)
			}
		})
	}

	t.Run("exactly 25 MP passes the size checks", func(t *testing.T) {
		_, err := enc.Encode(t.Context(), pngHeader(t, 5000, 5000), covers.Full)
		if errors.Is(err, covers.ErrTooLarge) {
			t.Fatalf("25 MP rejected: %v", err)
		}
	})
}
