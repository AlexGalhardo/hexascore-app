// Package backup dumps the database, encrypts it client-side and uploads it
// (SDD 11 §7.3, ADR 0014).
package backup

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/AlexGalhardo/vukem/crypt"
)

// FrameSize is the plaintext size of every frame but the last.
const FrameSize = 4 << 20

// ErrFrame reports a frame that does not open: tampered, reordered, from
// another object, or sealed with another key.
var ErrFrame = errors.New("backup: frame does not authenticate")

// ErrTruncated reports a stream that ends before its last frame, or data
// after it.
var ErrTruncated = errors.New("backup: stream truncated or extended")

// frameAD binds a frame to its object and position; the last frame is
// marked so dropping trailing frames is detected.
func frameAD(object string, n int, last bool) []byte {
	ad := "backup:" + object + ":" + strconv.Itoa(n)
	if last {
		ad += ":last"
	}

	return []byte(ad)
}

// FrameWriter seals plaintext in FrameSize frames, one token per line
// (crypt tokens are base64url, so they never contain a newline). Close
// seals the remaining bytes as the last frame, possibly empty.
type FrameWriter struct {
	w      io.Writer
	enc    *crypt.Encrypter
	object string
	size   int
	buf    []byte
	n      int
}

// NewFrameWriter returns a writer sealing frames for object into w.
func NewFrameWriter(w io.Writer, enc *crypt.Encrypter, object string) *FrameWriter {
	return &FrameWriter{w: w, enc: enc, object: object, size: FrameSize}
}

// Write implements io.Writer.
func (f *FrameWriter) Write(p []byte) (int, error) {
	f.buf = append(f.buf, p...)

	// Strictly greater: the bytes that may end the stream stay buffered, so
	// Close can mark them as the last frame.
	for len(f.buf) > f.size {
		if err := f.seal(f.buf[:f.size], false); err != nil {
			return 0, err
		}

		f.buf = append(f.buf[:0], f.buf[f.size:]...)
	}

	return len(p), nil
}

// Close seals the last frame. It does not close the underlying writer.
func (f *FrameWriter) Close() error {
	err := f.seal(f.buf, true)
	f.buf = nil

	return err
}

func (f *FrameWriter) seal(frame []byte, last bool) error {
	tok := f.enc.Seal(frame, frameAD(f.object, f.n, last))
	f.n++

	if _, err := io.WriteString(f.w, tok+"\n"); err != nil {
		return fmt.Errorf("backup: write frame: %w", err)
	}

	return nil
}

// FrameReader opens a FrameWriter stream. Read returns ErrFrame or
// ErrTruncated instead of io.EOF when the stream was tampered with.
type FrameReader struct {
	sc     *bufio.Scanner
	enc    *crypt.Encrypter
	object string
	n      int
	done   bool
	plain  bytes.Reader
}

// NewFrameReader returns a reader opening the frames of object from r.
func NewFrameReader(r io.Reader, enc *crypt.Encrypter, object string) *FrameReader {
	sc := bufio.NewScanner(r)
	// A token is base64 of header + nonce + frame + tag: 4/3 of FrameSize
	// plus a little; 2x leaves room without unbounded growth.
	sc.Buffer(make([]byte, 0, 64<<10), 2*FrameSize)

	return &FrameReader{sc: sc, enc: enc, object: object}
}

// Read implements io.Reader.
func (f *FrameReader) Read(p []byte) (int, error) {
	for f.plain.Len() == 0 {
		if err := f.next(); err != nil {
			return 0, err
		}
	}

	return f.plain.Read(p)
}

func (f *FrameReader) next() error {
	if !f.sc.Scan() {
		if err := f.sc.Err(); err != nil {
			return fmt.Errorf("backup: read frame: %w", err)
		}

		if f.done {
			return io.EOF
		}

		return ErrTruncated
	}

	if f.done {
		return ErrTruncated
	}

	tok := f.sc.Text()

	plain, err := f.enc.Open(tok, frameAD(f.object, f.n, false))
	if err != nil {
		if plain, err = f.enc.Open(tok, frameAD(f.object, f.n, true)); err != nil {
			return fmt.Errorf("%w: frame %d", ErrFrame, f.n)
		}

		f.done = true
	}

	f.n++
	f.plain.Reset(plain)

	return nil
}
