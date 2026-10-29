package catalog_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

const keyA = "0123456789abcdef0123456789abcdef"

func cursorCodec(t *testing.T, key string) *catalog.CursorCodec {
	t.Helper()
	c, err := catalog.NewCursorCodec([]byte(key))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCursorRoundTrip(t *testing.T) {
	t.Parallel()
	c := cursorCodec(t, keyA)
	in := catalog.Cursor{S: "-score", V: "70.0000", ID: "id-1", D: catalog.DirNext}
	got, err := c.Decode(c.Encode(in), "-score")
	if err != nil || got != in {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestCursorRejected(t *testing.T) {
	t.Parallel()
	c := cursorCodec(t, keyA)
	other := cursorCodec(t, "ffffffffffffffffffffffffffffffff")
	tok := c.Encode(catalog.Cursor{S: "-score", V: "1", ID: "id-1", D: catalog.DirNext})
	payload, sig, _ := strings.Cut(tok, ".")
	flipped := "A" + payload[1:]
	if payload[0] == 'A' {
		flipped = "B" + payload[1:]
	}
	noDir := c.Encode(catalog.Cursor{S: "-score", V: "1", ID: "id-1"})

	for name, tc := range map[string]struct {
		codec *catalog.CursorCodec
		token string
		sort  string
	}{
		"tampered payload": {c, flipped + "." + sig, "-score"},
		"no signature":     {c, payload, "-score"},
		"empty parts":      {c, ".", "-score"},
		"garbage":          {c, "!!!.???", "-score"},
		"other sort":       {c, tok, "-popularity"},
		"other direction":  {c, tok, "score"},
		"other key":        {other, tok, "-score"},
		"missing dir":      {c, noDir, "-score"},
	} {
		if _, err := tc.codec.Decode(tc.token, tc.sort); !errors.Is(err, catalog.ErrInvalidCursor) {
			t.Errorf("%s: err = %v, want ErrInvalidCursor", name, err)
		}
	}
}

func TestCursorNeedsKey(t *testing.T) {
	t.Parallel()
	if _, err := catalog.NewCursorCodec([]byte("short")); err == nil {
		t.Fatal("short key accepted")
	}
}
