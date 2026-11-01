package moderation

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Field kinds of the diff view (SDD 04 section 6): word diff for text, chips for relations, covers side by side.
const (
	KindText     = "text"
	KindRelation = "relation"
	KindCover    = "cover"
)

// WordOp is one token of a word diff: Op is eq, del or add.
type WordOp struct{ Op, Text string }

// FieldDiff is one changed field: the live value, the author's before and after, and the derived views.
type FieldDiff struct {
	Field, Kind            string
	Current, Before, After string
	Words                  []WordOp
	Add, Remove            []string
}

type relationOp struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type fieldChange struct {
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
	Add    []relationOp    `json:"add"`
	Remove []relationOp    `json:"remove"`
}

// BuildDiffs decodes a suggestion's changes object into per-field diffs sorted by field. current returns the live
// value of a scalar field; it may return nil.
func BuildDiffs(raw json.RawMessage, current func(field string) json.RawMessage) ([]FieldDiff, error) {
	var changes map[string]fieldChange
	if err := json.Unmarshal(raw, &changes); err != nil {
		return nil, fmt.Errorf("moderation: decode changes: %w", err)
	}

	out := make([]FieldDiff, 0, len(changes))

	for field, c := range changes {
		d := FieldDiff{Field: field, Kind: KindText}

		switch {
		case len(c.Add) > 0 || len(c.Remove) > 0:
			d.Kind = KindRelation
			d.Add, d.Remove = chips(c.Add), chips(c.Remove)
		default:
			d.Before, d.After = display(c.Before), display(c.After)
			if current != nil {
				d.Current = display(current(field))
			}

			if field == "cover" || strings.HasPrefix(field, "cover.") {
				d.Kind = KindCover
			} else {
				d.Words = WordDiff(d.Before, d.After)
			}
		}

		out = append(out, d)
	}

	slices.SortFunc(out, func(a, b FieldDiff) int { return strings.Compare(a.Field, b.Field) })

	return out, nil
}

func chips(ops []relationOp) []string {
	out := make([]string, 0, len(ops))

	for _, o := range ops {
		if o.Role != "" {
			out = append(out, o.ID+" ("+o.Role+")")
		} else {
			out = append(out, o.ID)
		}
	}

	return out
}

// display renders a JSON scalar as text: strings unquoted, null as empty.
func display(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}

	return string(raw)
}

// WordDiff is a whitespace-token longest-common-subsequence diff. Output is plain text; the template escapes it.
// shortcut: O(n*m) table, fine for titles and synopses capped at 300 and a few thousand characters.
func WordDiff(before, after string) []WordOp {
	a, b := strings.Fields(before), strings.Fields(after)
	lcs := make([][]int, len(a)+1)

	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}

	for i := range slices.Backward(a) {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}

	var ops []WordOp

	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			ops = append(ops, WordOp{"eq", a[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, WordOp{"del", a[i]})
			i++
		default:
			ops = append(ops, WordOp{"add", b[j]})
			j++
		}
	}

	for ; i < len(a); i++ {
		ops = append(ops, WordOp{"del", a[i]})
	}

	for ; j < len(b); j++ {
		ops = append(ops, WordOp{"add", b[j]})
	}

	return ops
}
