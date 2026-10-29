package search

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Input limits (SDD 07 NFR-07-6, §7.1).
const (
	minQueryRunes = 2
	maxQueryRunes = 100
	maxQueryTerms = 10
)

// ErrInvalidQuery marks input the caller must answer with 400 `invalid-query` (RFC 9457).
var ErrInvalidQuery = errors.New("invalid-query")

// ErrTimeout is returned when Postgres cancelled a statement on the search timeout; the route answers 503.
var ErrTimeout = errors.New("search timeout")

// Parsed is a query after sanitising (SDD 07 §7.1).
type Parsed struct {
	// Text is the cleaned query handed to websearch_to_tsquery (quotes, -exclusions and `or` intact).
	Text string
	// Match is Text without quotes, -terms and `or`: the input of the exact, prefix and trigram tiers.
	Match string
	// Excluded is the `-term` tokens as a websearch `or` list; trigram branches ignore operators, so the
	// searcher uses it to drop items whose text matches any of them (FR-07-3).
	Excluded string
	// ISBN13 is set when the whole query is an ISBN-13 (hyphens and spaces stripped).
	ISBN13 string
	// Empty queries answer 200 with no results; TooShort ones are flagged for the UI.
	Empty, TooShort bool
}

// ExclusionOnly reports queries made only of `-term` tokens: nothing to match, zero results.
func (p Parsed) ExclusionOnly() bool { return p.Text != "" && p.Match == "" && p.ISBN13 == "" }

// Parse sanitises raw user text: control characters become spaces, NFC, trim and collapse; more than 100 runes or
// 10 terms is ErrInvalidQuery; under 2 runes is TooShort.
func Parse(raw string) (Parsed, error) {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Cc, r) || r == utf8.RuneError {
			return ' '
		}

		return r
	}, raw)
	text := strings.Join(strings.Fields(norm.NFC.String(cleaned)), " ")

	switch n := utf8.RuneCountInString(text); {
	case n == 0:
		return Parsed{Empty: true}, nil
	case n > maxQueryRunes:
		return Parsed{}, fmt.Errorf("%w: query longer than %d characters", ErrInvalidQuery, maxQueryRunes)
	case n < minQueryRunes:
		return Parsed{Text: text, TooShort: true}, nil
	}

	if len(strings.Fields(text)) > maxQueryTerms {
		return Parsed{}, fmt.Errorf("%w: more than %d terms", ErrInvalidQuery, maxQueryTerms)
	}

	p := Parsed{Text: text, Match: matchText(text), Excluded: excludedText(text)}
	if digits := strings.NewReplacer("-", "", " ", "").Replace(text); isISBN13(digits) {
		p.ISBN13 = digits
	}

	return p, nil
}

// matchText drops websearch operators: `-term`, `-"phrase"` and `or` disappear, quote characters are removed.
func matchText(text string) string {
	var (
		words   []string
		cur     strings.Builder
		inQuote bool
		negated bool
	)

	flush := func() {
		if w := cur.String(); w != "" && !negated && (inQuote || !strings.EqualFold(w, "or")) {
			words = append(words, w)
		}
		cur.Reset()
	}

	atStart := true
	for _, r := range text {
		switch {
		case r == '"':
			if !inQuote && atStart {
				// `-"phrase"` keeps negated across the whole phrase.
				inQuote = true
				continue
			}
			flush()
			inQuote = false
			negated = false
			atStart = true
		case unicode.IsSpace(r) && !inQuote:
			flush()
			negated = false
			atStart = true
		case unicode.IsSpace(r):
			flush()
		case r == '-' && atStart && !inQuote:
			negated = true
		default:
			cur.WriteRune(r)
			atStart = false
		}
	}
	flush()

	return strings.Join(words, " ")
}

var negatedToken = regexp.MustCompile(`(?:^|\s)-("[^"]*"|\S+)`)

func excludedText(text string) string {
	var parts []string

	for _, m := range negatedToken.FindAllStringSubmatch(text, -1) {
		tok := strings.Trim(m[1], `"`)
		if tok == "" {
			continue
		}

		if strings.ContainsFunc(tok, unicode.IsSpace) {
			tok = `"` + tok + `"`
		}

		parts = append(parts, tok)
	}

	return strings.Join(parts, " or ")
}

func isISBN13(s string) bool {
	if len(s) != 13 || (!strings.HasPrefix(s, "978") && !strings.HasPrefix(s, "979")) {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
