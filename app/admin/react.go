package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
)

// react renders a React admin screen (ADR 0026) in place of a template. The screen's data is the same row structs
// the template used, with camelCase keys; the panel's one-shot flash messages travel with it and the response keeps
// the panel's no-store and frame headers.
func (h CatalogHandlers) react(c *web.Ctx, component string, props map[string]any) error {
	return renderReact(c, component, props)
}

func renderReact(c *web.Ctx, component string, props map[string]any) error {
	data, err := camelProps(props)
	if err != nil {
		return err
	}

	out, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("admin react %s: props are not an object", component)
	}

	out["flash"] = map[string]any{
		"success": flashText(c, "admin.success"),
		"warning": flashText(c, "admin.warning"),
	}

	c.SetHeader("Cache-Control", "no-store")
	c.SetHeader("X-Frame-Options", "DENY")

	return inertia.Render(c, component, inertia.Props(out))
}

func flashText(c *web.Ctx, key string) string {
	if s, ok := c.FlashValue(key).(string); ok {
		return s
	}

	return ""
}

// camelProps round-trips v through JSON and renames every object key to camelCase, so Go structs without JSON tags
// can feed React directly. Numbers keep their exact text (a bigint id must not turn into a float).
func camelProps(v any) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("admin react props: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	var out any
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("admin react props: %w", err)
	}

	return camelKeys(out), nil
}

func camelKeys(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, x := range t {
			m[camelKey(k)] = camelKeys(x)
		}

		return m
	case []any:
		s := make([]any, len(t))
		for i, x := range t {
			s[i] = camelKeys(x)
		}

		return s
	default:
		return v
	}
}

// camelKey lowers the leading capitals of a Go field name: ID is id, URL is url, TargetID is targetID and HTMLText is
// htmlText. Keys that already start in lower case stay as they are.
func camelKey(k string) string {
	runes := []rune(k)

	n := 0
	for n < len(runes) && unicode.IsUpper(runes[n]) {
		n++
	}

	switch {
	case n == 0:
		return k
	case n == len(runes):
		return strings.ToLower(k)
	case n == 1:
		return string(unicode.ToLower(runes[0])) + string(runes[1:])
	default:
		return strings.ToLower(string(runes[:n-1])) + string(runes[n-1:])
	}
}
