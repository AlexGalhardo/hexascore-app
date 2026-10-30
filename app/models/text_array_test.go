package models

import "testing"

func TestTextArrayRoundTrip(t *testing.T) {
	samples := []TextArray{
		{},
		{"https://example.com"},
		{"https://a.example/a,b", "https://b.example"},
		{`say "hi"`, `a\b`},
		{"NULL", ""},
	}
	for _, in := range samples {
		literal, err := in.Value()
		if err != nil {
			t.Fatal(err)
		}
		var got TextArray
		if err := got.Scan(literal); err != nil {
			t.Fatalf("scan %q: %v", literal, err)
		}
		if len(got) != len(in) {
			t.Fatalf("len %q = %d, want %d", literal, len(got), len(in))
		}
		for i := range in {
			if got[i] != in[i] {
				t.Fatalf("elem %d of %q = %q, want %q", i, literal, got[i], in[i])
			}
		}
	}
}

func TestTextArrayRejectsNullElement(t *testing.T) {
	var got TextArray
	if err := got.Scan("{NULL}"); err == nil {
		t.Fatal("expected null element error")
	}
}

func TestTextArrayScanBytes(t *testing.T) {
	var got TextArray
	if err := got.Scan([]byte(`{https://example.com}`)); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "https://example.com" {
		t.Fatalf("got %#v", got)
	}
}
