package covers_test

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
)

func TestSecurity_A04_FingerprintRejectsBombsBeforeDecoding(t *testing.T) {
	hdr := pngBytes(t, gradient(2, 2, 255))[:33]
	binary.BigEndian.PutUint32(hdr[16:], 30000)
	binary.BigEndian.PutUint32(hdr[20:], 30000)
	binary.BigEndian.PutUint32(hdr[29:], crc32.ChecksumIEEE(hdr[12:29]))

	if _, _, _, err := covers.Fingerprint(hdr); !errors.Is(err, covers.ErrTooLarge) {
		t.Fatalf("Fingerprint error = %v, want ErrTooLarge", err)
	}
}
