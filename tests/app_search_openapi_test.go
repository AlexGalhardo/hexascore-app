package tests

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

// TestAppSearchOpenAPIQuery: the app client is generated from openapi/app-v1.json, so the search reads must
// declare their query parameters there or the client cannot send q (scripts/openapi-check.sh keeps the golden
// equal to the export).
func TestAppSearchOpenAPIQuery(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../openapi/app-v1.json")
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
			Parameters  []struct {
				Name string `json:"name"`
				In   string `json:"in"`
			} `json:"parameters"`
			Responses map[string]struct {
				Content map[string]struct {
					Schema map[string]any `json:"schema"`
				} `json:"content"`
			} `json:"responses"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}

	for path, want := range map[string][]string{
		"/app/v1/search":         {"q", "type", "locale", "page"},
		"/app/v1/search/suggest": {"q", "type", "locale"},
	} {
		op := doc.Paths[path]["get"]

		var got []string
		for _, p := range op.Parameters {
			if p.In == "query" {
				got = append(got, p.Name)
			}
		}

		for _, name := range want {
			if !slices.Contains(got, name) {
				t.Errorf("%s: query parameter %q missing (got %v)", path, name, got)
			}
		}

		if len(op.Responses["200"].Content["application/json"].Schema) == 0 {
			t.Errorf("%s: 200 has no JSON schema", path)
		}
	}
}
