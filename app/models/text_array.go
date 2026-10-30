package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// TextArray scans a Postgres text[] column. The driver returns the array
// literal, not JSON. A null element is refused: source URL lists are not
// nullable elements.
type TextArray []string

// Value encodes the slice as a Postgres text[] literal.
func (a TextArray) Value() (driver.Value, error) {
	return formatTextArray(a), nil
}

// Scan implements sql.Scanner.
func (a *TextArray) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*a = nil
		return nil
	case []string:
		*a = append(TextArray(nil), v...)
		return nil
	case string:
		parsed, err := parseTextArray(v)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	case []byte:
		parsed, err := parseTextArray(string(v))
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	default:
		return fmt.Errorf("models: cannot scan %T into text[]", src)
	}
}

func formatTextArray(values []string) string {
	if len(values) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, v := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(quoteArrayElem(v))
	}
	b.WriteByte('}')
	return b.String()
}

func quoteArrayElem(v string) string {
	if v == "" || strings.EqualFold(v, "NULL") || strings.ContainsAny(v, "{},\"\\ \t\n\r") {
		var b strings.Builder
		b.WriteByte('"')
		for i := 0; i < len(v); i++ {
			if v[i] == '"' || v[i] == '\\' {
				b.WriteByte('\\')
			}
			b.WriteByte(v[i])
		}
		b.WriteByte('"')
		return b.String()
	}
	return v
}

func parseTextArray(s string) (TextArray, error) {
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return nil, fmt.Errorf("models: invalid text[] %q", s)
	}
	body := s[1 : len(s)-1]
	if body == "" {
		return TextArray{}, nil
	}
	var out TextArray
	for i := 0; i < len(body); {
		if body[i] == '"' {
			var elem strings.Builder
			i++
			closed := false
			for i < len(body) {
				c := body[i]
				if c == '\\' {
					if i+1 >= len(body) {
						return nil, fmt.Errorf("models: truncated text[] escape")
					}
					elem.WriteByte(body[i+1])
					i += 2
					continue
				}
				if c == '"' {
					i++
					closed = true
					break
				}
				elem.WriteByte(c)
				i++
			}
			if !closed {
				return nil, fmt.Errorf("models: unterminated text[] quote")
			}
			out = append(out, elem.String())
		} else {
			j := i
			for j < len(body) && body[j] != ',' {
				j++
			}
			raw := body[i:j]
			if raw == "NULL" {
				return nil, fmt.Errorf("models: null element in text[]")
			}
			out = append(out, raw)
			i = j
		}
		if i < len(body) {
			if body[i] != ',' {
				return nil, fmt.Errorf("models: invalid text[] %q", s)
			}
			i++
			if i == len(body) {
				return nil, fmt.Errorf("models: trailing comma in text[]")
			}
		}
	}
	return out, nil
}
