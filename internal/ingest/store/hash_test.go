package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

func TestCanonicalHashIgnoresKeyOrder(t *testing.T) {
	a, err := store.CanonicalHash([]byte(`{"b":[1,{"y":2,"x":1}],"a":"é"}`))
	if err != nil {
		t.Fatal(err)
	}

	b, err := store.CanonicalHash([]byte(`{ "a": "é", "b": [1, {"x": 1, "y": 2}] }`))
	if err != nil {
		t.Fatal(err)
	}

	if a != b {
		t.Errorf("key order changed the hash: %s vs %s", a, b)
	}

	c, _ := store.CanonicalHash([]byte(`{"a":"é","b":[{"x":1,"y":2},1]}`))
	if a == c {
		t.Error("array order must change the hash")
	}
}

// Expected values come from legacy/webscrappings/common/db.py canonical_hash over the same files.
func TestCanonicalHashMatchesLegacyPython(t *testing.T) {
	cases := map[string]string{
		"wikidata_like": "ca8b57c6d20e9f56ca0fa138a06f09eb1bb20832c941566fe30066fd6716f7ee",
		"text_escapes":  "c73f2db3f94cf3bad5530fe21eb3b6fdfbcb168d7215794f6e7570f2ccf11a8f",
		"numbers":       "63d60d860bb4536df4cbfdca801722591d73bb2d6adc3301126eb61498bf5b90",
	}

	for name, want := range cases {
		raw, err := os.ReadFile(filepath.Join("testdata", name+".json"))
		if err != nil {
			t.Fatal(err)
		}

		got, err := store.CanonicalHash(raw)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if got != want {
			t.Errorf("%s: hash %s, want %s", name, got, want)
		}
	}
}

func TestCanonicalHashRejectsBadJSON(t *testing.T) {
	for _, in := range []string{``, `{`, `{"a":1} {"b":2}`, `1e999`} {
		if _, err := store.CanonicalHash([]byte(in)); err == nil {
			t.Errorf("%q: want error", in)
		}
	}
}
