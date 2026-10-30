package community

import (
	"encoding/json"
	"regexp"

	"github.com/AlexGalhardo/vukem/z"
	"golang.org/x/text/unicode/norm"
)

// SuggestionInput is one community suggestion before it is stored.
type SuggestionInput struct {
	Kind       string
	TargetType string
	OtherType  string
	Locale     string
	Note       string
	SourceURLs []string
	Changes    map[string]FieldChange
}

// FieldChange is one scalar edit (Before/After) or one relation edit (Add/Remove).
type FieldChange struct {
	Before json.RawMessage
	After  json.RawMessage
	Add    []RelationOp
	Remove []RelationOp
}

// RelationOp adds or removes one entity on a relation field.
type RelationOp struct {
	ID   string
	Role string
}

// ValidationError is a 422 from suggestion field validation.
type ValidationError struct {
	Status int
	Field  string
	Msg    string
}

// Error implements error.
func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Field == "" {
		return e.Msg
	}

	return e.Field + ": " + e.Msg
}

func invalid(field, msg string) *ValidationError {
	return &ValidationError{Status: 422, Field: field, Msg: msg}
}

// Schemas are the SDD 04 §7.2 rules. z exports them as JSON Schema (ADR 0009).
var (
	kindSchema        = z.Enum("edit", "create", "merge")
	targetSchema      = z.Enum("movie", "series", "book", "game")
	localeSchema      = z.Enum("en", "pt-BR", "es")
	noteSchema        = z.String().Max(1000)
	titleSchema       = z.String().Preprocess(norm.NFC.String).Min(1).Max(300).Refine(rejectLineControls)
	synopsisSchema    = z.String().Preprocess(norm.NFC.String).Min(50).Max(2000)
	precisionSchema   = z.Enum("day", "month", "year")
	sourceURLSchema   = z.String().Max(2048).URL().Refine(allowedSourceURL)
	coverSchema       = z.String().Max(2048).Refine(allowedCover)
	wikidataSchema    = z.String().Regex(regexp.MustCompile(`^Q\d+$`))
	digitsSchema      = z.String().Regex(regexp.MustCompile(`^\d+$`))
	openLibrarySchema = z.String().Regex(regexp.MustCompile(`^OL\d+[WMA]$`))
	idSchema          = z.String().Min(1).Max(64)
)

// shortenerHosts is the SDD 04 §7.2 denylist. The SDD requires the refusal
// and does not name hosts; these are public shorteners that hide the cited page.
var shortenerHosts = map[string]struct{}{
	"bit.ly": {}, "j.mp": {}, "tinyurl.com": {}, "t.co": {}, "goo.gl": {},
	"ow.ly": {}, "is.gd": {}, "buff.ly": {}, "cutt.ly": {}, "rb.gy": {},
	"rebrand.ly": {}, "tiny.cc": {}, "shorturl.at": {}, "t.ly": {}, "v.gd": {},
	"clck.ru": {}, "git.io": {}, "amzn.to": {}, "lnkd.in": {}, "db.tt": {},
	"adf.ly": {}, "soo.gd": {}, "s.id": {}, "bl.ink": {},
}
