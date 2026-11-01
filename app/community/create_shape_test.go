package community

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCheckCreateShape(t *testing.T) {
	t.Parallel()

	ch := func(v string) FieldChange { return FieldChange{After: json.RawMessage(`"` + v + `"`)} }
	const work = "00000000-0000-0000-0000-0000000000cb"

	cases := []struct {
		name    string
		target  string
		domain  string
		changes map[string]FieldChange
		field   string
	}{
		{"item ok", "item", "movie", map[string]FieldChange{"title": ch("A")}, ""},
		{
			"item refuses entity_kind",
			"item",
			"movie",
			map[string]FieldChange{"title": ch("A"), "entity_kind": ch("person")},
			"entity_kind",
		},
		{"entity ok", "entity", "movie", map[string]FieldChange{"title": ch("A"), "entity_kind": ch("person")}, ""},
		{"entity needs kind", "entity", "movie", map[string]FieldChange{"title": ch("A")}, "entity_kind"},
		{
			"entity refuses synopsis",
			"entity",
			"movie",
			map[string]FieldChange{"title": ch("A"), "entity_kind": ch("genre"), "synopsis.en": ch("x")},
			"synopsis.en",
		},
		{
			"edition ok",
			"edition",
			"book",
			map[string]FieldChange{
				"title":               ch("A"),
				"work_id":             ch(work),
				"language":            ch("en"),
				"external_ids.isbn13": ch("9780306406157"),
			},
			"",
		},
		{
			"edition needs book",
			"edition",
			"movie",
			map[string]FieldChange{
				"title":               ch("A"),
				"work_id":             ch(work),
				"language":            ch("en"),
				"external_ids.isbn13": ch("9780306406157"),
			},
			"target_type",
		},
		{
			"edition needs work",
			"edition",
			"book",
			map[string]FieldChange{"title": ch("A"), "language": ch("en"), "external_ids.isbn13": ch("9780306406157")},
			"work_id",
		},
		{
			"edition needs language",
			"edition",
			"book",
			map[string]FieldChange{"title": ch("A"), "work_id": ch(work), "external_ids.isbn13": ch("9780306406157")},
			"language",
		},
		{
			"edition needs an external id",
			"edition",
			"book",
			map[string]FieldChange{"title": ch("A"), "work_id": ch(work), "language": ch("en")},
			"external_ids",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in := SuggestionInput{Kind: "create", TargetType: tc.domain, Changes: tc.changes}
			err := checkCreate(Target{Kind: tc.target}, in)
			if tc.field == "" {
				if err != nil {
					t.Fatalf("err = %v", err)
				}
				return
			}
			var ve *ValidationError
			if !errors.As(err, &ve) || ve.Field != tc.field {
				t.Fatalf("err = %#v, want field %s", err, tc.field)
			}
		})
	}
}

func TestValidateCreateOnlyFields(t *testing.T) {
	t.Parallel()

	base := SuggestionInput{Kind: "create", TargetType: "book", Locale: "en"}
	for field, value := range map[string]string{
		"entity_kind": "country", "work_id": "not-a-uuid", "language": "english",
	} {
		in := base
		in.Changes = map[string]FieldChange{
			"title": {After: json.RawMessage(`"A"`)}, field: {After: json.RawMessage(`"` + value + `"`)},
		}
		var ve *ValidationError
		if err := Validate(in, validateNow); !errors.As(err, &ve) || ve.Field != field {
			t.Fatalf("%s=%s: err = %#v", field, value, err)
		}
	}

	// On an edit the create-only fields are unknown.
	edit := SuggestionInput{
		Kind: "edit", TargetType: "movie", Locale: "en",
		Changes: map[string]FieldChange{"entity_kind": {After: json.RawMessage(`"person"`)}},
	}
	var ve *ValidationError
	if err := Validate(edit, validateNow); !errors.As(err, &ve) || ve.Field != "entity_kind" {
		t.Fatalf("edit err = %#v", err)
	}
}
