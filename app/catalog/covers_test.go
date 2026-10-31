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

func TestCoverAbsolute(t *testing.T) {
	c := &Cover{Thumb: "/covers/a-t.webp", Full: "/covers/a.webp"}

	got := c.WithBase("https://bucket.example")
	if got.Thumb != "https://bucket.example/covers/a-t.webp" || got.Full != "https://bucket.example/covers/a.webp" {
		t.Fatalf("absolute renditions: %+v", got)
	}

	if c.Thumb != "/covers/a-t.webp" {
		t.Fatalf("WithBase mutated its receiver: %+v", c)
	}

	if c.WithBase("") != c || (*Cover)(nil).WithBase("https://x") != nil {
		t.Fatal("empty base and nil cover must pass through")
	}

	if CoversBase(" https://bucket.example/ ") != "https://bucket.example" || Absolute("https://x", "") != "" {
		t.Fatal("base trimming or empty path")
	}
}
