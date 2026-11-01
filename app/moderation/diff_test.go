package moderation

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWordDiff(t *testing.T) {
	ops := WordDiff("the quick fox", "the slow fox jumps")
	var got []string
	for _, o := range ops {
		got = append(got, o.Op+":"+o.Text)
	}
	want := "eq:the del:quick add:slow eq:fox add:jumps"
	if strings.Join(got, " ") != want {
		t.Fatalf("WordDiff = %q, want %q", strings.Join(got, " "), want)
	}
}

func TestBuildDiffs(t *testing.T) {
	raw := json.RawMessage(`{
"title.en":{"before":"Old","after":"New <b>"},
"genres":{"add":[{"id":"g1","role":""}],"remove":[{"id":"g2","role":""}]},
"cover":{"before":null,"after":"community/covers/x.webp"}}`)
	current := func(field string) json.RawMessage {
		if field == "title.en" {
			return json.RawMessage(`"Old"`)
		}

		return json.RawMessage(`null`)
	}

	diffs, err := BuildDiffs(raw, current)
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]FieldDiff{}
	for _, d := range diffs {
		by[d.Field] = d
	}
	if d := by["title.en"]; d.Kind != KindText || d.Current != "Old" || d.After != "New <b>" || len(d.Words) == 0 {
		t.Fatalf("title diff: %+v", d)
	}
	if d := by["genres"]; d.Kind != KindRelation || len(d.Add) != 1 || len(d.Remove) != 1 {
		t.Fatalf("relation diff: %+v", d)
	}
	if d := by["cover"]; d.Kind != KindCover || d.After != "community/covers/x.webp" {
		t.Fatalf("cover diff: %+v", d)
	}
	if diffs[0].Field != "cover" {
		t.Fatalf("diffs not sorted by field: %s first", diffs[0].Field)
	}
}
