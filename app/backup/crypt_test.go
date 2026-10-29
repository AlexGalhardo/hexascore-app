package backup

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/crypt"
)

const object = "backups/postgres/daily/2026/10/29/hexascore-test-20261029T0900Z.dump.enc"

func encrypter(t *testing.T) *crypt.Encrypter {
	t.Helper()

	enc, err := crypt.New(bytes.Repeat([]byte{9}, 32))
	if err != nil {
		t.Fatal(err)
	}

	return enc
}

// seal frames plain with a small frame size so tests stay fast; the
// production size is exercised by TestFrameEncryptRoundTrip.
func seal(t *testing.T, enc *crypt.Encrypter, plain []byte, size int) []string {
	t.Helper()

	var out bytes.Buffer

	w := NewFrameWriter(&out, enc, object)
	w.size = size

	if _, err := w.Write(plain); err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	return strings.SplitAfter(out.String(), "\n")[:w.n]
}

func open(enc *crypt.Encrypter, frames []string, obj string) ([]byte, error) {
	return io.ReadAll(NewFrameReader(strings.NewReader(strings.Join(frames, "")), enc, obj))
}

func TestFrameEncryptRoundTrip(t *testing.T) {
	t.Parallel()

	enc := encrypter(t)

	for _, n := range []int{0, 1, FrameSize, FrameSize + 1, 2*FrameSize + 123} {
		plain := make([]byte, n)
		_, _ = rand.Read(plain)

		var out bytes.Buffer

		w := NewFrameWriter(&out, enc, object)

		// Odd write sizes cross frame boundaries the way an io.Copy does.
		for rest := plain; len(rest) > 0; {
			k := min(len(rest), 1<<20+7)
			if _, err := w.Write(rest[:k]); err != nil {
				t.Fatal(err)
			}

			rest = rest[k:]
		}

		if err := w.Close(); err != nil {
			t.Fatal(err)
		}

		if bytes.Contains(out.Bytes(), plain[:min(n, 64)]) && n >= 64 {
			t.Fatalf("n=%d: plaintext visible in ciphertext", n)
		}

		got, err := io.ReadAll(NewFrameReader(&out, enc, object))
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}

		if !bytes.Equal(got, plain) {
			t.Fatalf("n=%d: round trip mismatch (%d bytes back)", n, len(got))
		}
	}
}

func TestFrameReorderRejected(t *testing.T) {
	t.Parallel()

	enc := encrypter(t)
	frames := seal(t, enc, bytes.Repeat([]byte("abcdefgh"), 4), 8)

	if len(frames) != 4 {
		t.Fatalf("frames = %d, want 4", len(frames))
	}

	frames[0], frames[1] = frames[1], frames[0]
	if _, err := open(enc, frames, object); !errors.Is(err, ErrFrame) {
		t.Fatalf("swapped frames: err = %v, want ErrFrame", err)
	}

	frames = seal(t, enc, bytes.Repeat([]byte("abcdefgh"), 4), 8)
	if _, err := open(enc, frames, "backups/postgres/daily/other.dump.enc"); !errors.Is(err, ErrFrame) {
		t.Fatalf("frames of another object: err = %v, want ErrFrame", err)
	}

	other, err := crypt.New(bytes.Repeat([]byte{8}, 32))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := open(other, frames, object); !errors.Is(err, ErrFrame) {
		t.Fatalf("wrong key: err = %v, want ErrFrame", err)
	}
}

func TestFrameTruncationRejected(t *testing.T) {
	t.Parallel()

	enc := encrypter(t)
	frames := seal(t, enc, bytes.Repeat([]byte("abcdefgh"), 4), 8)

	if _, err := open(enc, frames[:len(frames)-1], object); !errors.Is(err, ErrTruncated) {
		t.Fatalf("last frame dropped: err = %v, want ErrTruncated", err)
	}

	if _, err := open(enc, append(slices.Clip(frames), frames[0]), object); !errors.Is(err, ErrTruncated) {
		t.Fatalf("frame appended after last: err = %v, want ErrTruncated", err)
	}

	cut := append([]string{}, frames...)
	cut[1] = cut[1][:len(cut[1])/2] + "\n"

	if _, err := open(enc, cut, object); !errors.Is(err, ErrFrame) {
		t.Fatalf("frame cut in half: err = %v, want ErrFrame", err)
	}
}
