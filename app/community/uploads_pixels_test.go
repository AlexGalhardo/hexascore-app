package community

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"testing"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/storage"
)

// headerPNG is a PNG whose IHDR declares w x h with no pixel data: DecodeConfig reads it, a decode would fail.
func headerPNG(t *testing.T, w, h uint32) []byte {
	t.Helper()

	var b bytes.Buffer
	if err := png.Encode(&b, image.NewGray(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}

	hdr := b.Bytes()[:33]
	binary.BigEndian.PutUint32(hdr[16:], w)
	binary.BigEndian.PutUint32(hdr[20:], h)
	binary.BigEndian.PutUint32(hdr[29:], crc32.ChecksumIEEE(hdr[12:29]))

	return hdr
}

func TestSecurity_A04_CoverUploadRejectsDecompressionBombs(t *testing.T) {
	t.Parallel()

	disk := storage.NewMemory("")

	_, err := SaveCover(t.Context(), disk, "cover.png", bytes.NewReader(headerPNG(t, 30000, 30000)))
	if _, ok := errors.AsType[*errs.ValidationError](err); !ok {
		t.Fatalf("30000x30000 cover error = %v, want a validation error", err)
	}
}

func TestCoverUploadKeepsNormalSizes(t *testing.T) {
	t.Parallel()

	disk := storage.NewMemory("")

	c, err := SaveCover(t.Context(), disk, "cover.png", bytes.NewReader(headerPNG(t, 600, 900)))
	if err != nil || c.Width != 600 || c.Height != 900 {
		t.Fatalf("600x900 cover = %+v, %v", c, err)
	}
}
