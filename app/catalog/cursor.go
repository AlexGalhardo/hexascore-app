package catalog

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidCursor is returned for any cursor that is malformed, forged,
// signed with another key or issued for another sort; handlers answer 400
// with problem code invalid-cursor.
var ErrInvalidCursor = errors.New("invalid-cursor")

// Cursor directions: DirNext continues the sort order, DirPrev walks it backwards.
const (
	DirNext = "next"
	DirPrev = "prev"
)

// Cursor is the decoded position: the sort key, the last row's sort value
// (text, cast back in SQL), its id and the walking direction.
type Cursor struct {
	S  string `json:"s"`
	V  string `json:"v"`
	ID string `json:"id"`
	D  string `json:"d"`
}

// CursorCodec signs and verifies cursors with HMAC-SHA256 (VF-005: vukem's
// CursorPaginate pages by primary key only).
type CursorCodec struct{ key []byte }

// NewCursorCodec needs a key of at least 32 bytes; derive it from APP_KEY with
// App.DeriveKey("catalog.cursor", 32).
func NewCursorCodec(key []byte) (*CursorCodec, error) {
	if len(key) < 32 {
		return nil, fmt.Errorf("catalog: cursor key must be at least 32 bytes, got %d", len(key))
	}
	return &CursorCodec{key: key}, nil
}

var b64 = base64.RawURLEncoding

// Encode returns payload.signature, both base64url. The sort key is part of
// the signed payload, which binds the cursor to its sort and direction.
func (c *CursorCodec) Encode(cur Cursor) string {
	raw, _ := json.Marshal(cur) // plain strings: cannot fail
	p := b64.EncodeToString(raw)
	return p + "." + b64.EncodeToString(c.mac(p))
}

// Decode verifies token for the sort key sortKey. Every failure is ErrInvalidCursor.
func (c *CursorCodec) Decode(token, sortKey string) (Cursor, error) {
	p, sig, ok := strings.Cut(token, ".")
	if !ok || p == "" || sig == "" {
		return Cursor{}, ErrInvalidCursor
	}
	got, err := b64.DecodeString(sig)
	if err != nil || !hmac.Equal(got, c.mac(p)) {
		return Cursor{}, ErrInvalidCursor
	}
	raw, err := b64.DecodeString(p)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	var cur Cursor
	if err := json.Unmarshal(raw, &cur); err != nil || cur.S != sortKey || cur.ID == "" ||
		(cur.D != DirNext && cur.D != DirPrev) {
		return Cursor{}, ErrInvalidCursor
	}
	return cur, nil
}

func (c *CursorCodec) mac(payload string) []byte {
	m := hmac.New(sha256.New, c.key)
	m.Write([]byte(payload))
	return m.Sum(nil)
}
