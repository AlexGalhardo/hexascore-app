package admin

import (
	"encoding/json"
	"testing"
)

func TestCamelKey(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"ID":               "id",
		"URL":              "url",
		"Job":              "job",
		"TargetID":         "targetID",
		"OtherID":          "otherID",
		"HTMLText":         "htmlText",
		"canRun":           "canRun",
		"QID":              "qid",
		"RawName":          "rawName",
		"OldestTMDB":       "oldestTMDB",
		"":                 "",
		"occurred_at":      "occurredAt",
		"reverts_audit_id": "revertsAuditId",
		"target_id":        "targetId",
	} {
		if got := camelKey(in); got != want {
			t.Errorf("camelKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCamelPropsNestsAndKeepsBigNumbers(t *testing.T) {
	t.Parallel()

	type row struct {
		ID      int64
		OtherID string
		Items   []struct{ Name string }
	}

	r := row{ID: 9007199254740993, OtherID: "x"}
	r.Items = append(r.Items, struct{ Name string }{"a"})

	got, err := camelProps(map[string]any{"Rows": []row{r}})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	const want = `{"rows":[{"id":9007199254740993,"items":[{"name":"a"}],"otherID":"x"}]}`
	if string(raw) != want {
		t.Fatalf("props = %s, want %s", raw, want)
	}
}
