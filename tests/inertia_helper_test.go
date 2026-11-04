//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"regexp"
	"testing"
)

var inertiaScript = regexp.MustCompile(`(?s)<script data-page="app" type="application/json">(.*?)</script>`)

// inertiaPage returns the component and the decoded props of an Inertia page rendered as HTML (ADR 0026: the admin
// screens are React, so their data is in the page script, not in markup).
func inertiaPage(t *testing.T, body string) (string, map[string]any) {
	t.Helper()

	m := inertiaScript.FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no Inertia page in the response: %.400s", body)
	}

	var page struct {
		Component string         `json:"component"`
		Props     map[string]any `json:"props"`
	}

	if err := json.Unmarshal([]byte(m[1]), &page); err != nil {
		t.Fatalf("Inertia page is not JSON: %v", err)
	}

	return page.Component, page.Props
}

// inertiaProp checks the page component and returns one prop as unescaped JSON text, so a test can assert that a
// value reached the screen without matching the JSON escaping of "<" or "&".
func inertiaProp(t *testing.T, body, component, key string) string {
	t.Helper()

	got, props := inertiaPage(t, body)
	if got != component {
		t.Fatalf("page component = %q, want %q", got, component)
	}

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(props[key]); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}
