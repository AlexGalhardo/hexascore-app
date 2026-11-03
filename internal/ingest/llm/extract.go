package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Extraction is the answer the prompt asks for: only facts the text states outright, null otherwise.
type Extraction struct {
	ReleaseDate    *string `json:"release_date"`
	RuntimeMinutes *int    `json:"runtime_minutes"`
	PageCount      *int    `json:"page_count"`
}

// Bounds of the values; the runtime and page limits are the CHECK constraints of items.
const (
	minReleaseYear = 1850
	maxYearsAhead  = 3
	maxRuntime     = 6000
	maxPages       = 20000
)

// ParseExtraction validates the model's text against the schema: one JSON object with only the known keys, a full ISO
// date, whole numbers in range, nothing after it. The model's output is untrusted (it read untrusted text), so
// anything else is an error and the call is recorded as invalid.
func ParseExtraction(raw string, now time.Time) (Extraction, error) {
	text := strings.TrimSpace(raw)
	if rest, ok := strings.CutPrefix(text, "```"); ok {
		_, rest, _ = strings.Cut(rest, "\n") // drop the "json" tag of the fence
		text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(rest), "```"))
	}

	dec := json.NewDecoder(strings.NewReader(text))
	dec.DisallowUnknownFields()

	var e Extraction
	if err := dec.Decode(&e); err != nil {
		return Extraction{}, fmt.Errorf("llm: answer is not the expected object: %w", err)
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Extraction{}, errors.New("llm: text after the object")
	}

	if e.ReleaseDate != nil {
		d, err := time.Parse(time.DateOnly, *e.ReleaseDate)
		if err != nil || d.Year() < minReleaseYear || d.Year() > now.Year()+maxYearsAhead {
			return Extraction{}, fmt.Errorf("llm: release_date %q is not a plausible date", *e.ReleaseDate)
		}
	}

	if e.RuntimeMinutes != nil && (*e.RuntimeMinutes < 1 || *e.RuntimeMinutes > maxRuntime) {
		return Extraction{}, fmt.Errorf("llm: runtime_minutes %d is out of range", *e.RuntimeMinutes)
	}

	if e.PageCount != nil && (*e.PageCount < 1 || *e.PageCount > maxPages) {
		return Extraction{}, fmt.Errorf("llm: page_count %d is out of range", *e.PageCount)
	}

	return e, nil
}

// Empty reports whether the model found nothing.
func (e Extraction) Empty() bool {
	return e.ReleaseDate == nil && e.RuntimeMinutes == nil && e.PageCount == nil
}

// jsonValue renders a value for field_reviews.value.
func jsonValue(v any) json.RawMessage {
	b, _ := json.Marshal(v)

	return bytes.Clone(b)
}
