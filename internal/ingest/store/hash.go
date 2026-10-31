// Package store is the ingestion record store (SDD 06 section 5): canonical hashing and the upsert that
// keeps history in the same statement.
package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
)

// CanonicalHash returns the SHA-256 hex of payload serialised as the legacy Python
// json.dumps(sort_keys=True, separators=(",", ":"), ensure_ascii=False) (common/db.py canonical_hash), so a
// row hashed by either side compares equal. payload is any valid JSON document.
func CanonicalHash(payload []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return "", fmt.Errorf("canonical hash: %w", err)
	}

	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return "", errors.New("canonical hash: trailing data after the JSON value")
	}

	var b bytes.Buffer
	if err := writeCanonical(&b, v); err != nil {
		return "", err
	}

	sum := sha256.Sum256(b.Bytes())

	return hex.EncodeToString(sum[:]), nil
}

func writeCanonical(b *bytes.Buffer, v any) error {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case string:
		writeString(b, x)
	case json.Number:
		s, err := pythonNumber(x.String())
		if err != nil {
			return err
		}

		b.WriteString(s)
	case []any:
		b.WriteByte('[')

		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}

			if err := writeCanonical(b, e); err != nil {
				return err
			}
		}

		b.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}

		// UTF-8 byte order equals Python's code point order.
		slices.Sort(keys)
		b.WriteByte('{')

		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}

			writeString(b, k)
			b.WriteByte(':')

			if err := writeCanonical(b, x[k]); err != nil {
				return err
			}
		}

		b.WriteByte('}')
	default:
		return fmt.Errorf("canonical hash: unexpected type %T", v)
	}

	return nil
}

// writeString escapes like Python with ensure_ascii=False: the quote, the backslash and control characters
// only. <, >, & and U+2028 stay raw UTF-8 (Go's encoding/json would escape them).
func writeString(b *bytes.Buffer, s string) {
	b.WriteByte('"')

	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\b':
			b.WriteString(`\b`)
		case r == '\f':
			b.WriteString(`\f`)
		case r < 0x20:
			fmt.Fprintf(b, `\u%04x`, r)
		default:
			b.WriteRune(r)
		}
	}

	b.WriteByte('"')
}

// pythonNumber renders a JSON number the way Python re-serialises a parsed one: integers keep their digits
// (any size); the rest become a float with repr formatting (1e16 -> 1e+16, 1E-5 -> 1e-05, 100.0 -> 100.0).
func pythonNumber(lit string) (string, error) {
	if !strings.ContainsAny(lit, ".eE") {
		if lit == "-0" {
			return "0", nil
		}

		return lit, nil
	}

	f, err := strconv.ParseFloat(lit, 64)
	if err != nil || math.IsInf(f, 0) {
		return "", fmt.Errorf("canonical hash: number %q is not a finite float64", lit)
	}

	// Shortest round-trip digits, then Python's layout: exponent form when exp < -4 or exp >= 16.
	mant, expStr, _ := strings.Cut(strconv.FormatFloat(f, 'e', -1, 64), "e")
	exp, _ := strconv.Atoi(expStr)

	if exp < -4 || exp >= 16 {
		return fmt.Sprintf("%se%+03d", mant, exp), nil
	}

	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}

	return s, nil
}
