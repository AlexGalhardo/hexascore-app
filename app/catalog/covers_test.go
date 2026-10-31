package catalog

import (
	"database/sql"
	"testing"
)

func TestCoverURL(t *testing.T) {
	if got := NewCover(sql.NullString{}, sql.NullString{}); got != nil {
		t.Fatalf("no keys: got %+v, want nil placeholder", got)
	}

	if got := CoverURL("../secret"); got != "" {
		t.Fatalf("traversal key: got %q, want empty", got)
	}

	if got := CoverURL("covers/a-1234.webp"); got != "/covers/a-1234.webp" {
		t.Fatalf("same-origin key: got %q", got)
	}

	want := &Cover{Thumb: "/covers/a-1-t.webp", Full: "/covers/a-1.webp"}
	got := CoverFromThumbKey(new("covers/a-1-t.webp"))
	if *got != *want {
		t.Fatalf("thumb-derived URLs: got %+v, want %+v", got, want)
	}

	only := NewCover(sql.NullString{}, sql.NullString{String: "covers/b-t.webp", Valid: true})
	if only.Full != only.Thumb {
		t.Fatalf("missing full key must fall back to the thumb: %+v", only)
	}
}
