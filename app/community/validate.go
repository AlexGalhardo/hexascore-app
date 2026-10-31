package community

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/AlexGalhardo/vukem/z"
	"golang.org/x/text/unicode/norm"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// Validate checks one suggestion against SDD 04 §7.2. now is the clock for
// the release-date ceiling (now + 10 years). Text is NFC-normalised before
// length checks and before/after comparison.
func Validate(in SuggestionInput, now time.Time) error {
	if _, err := kindSchema.Parse(in.Kind); err != nil {
		return invalid("kind", "invalid")
	}
	if _, err := targetSchema.Parse(in.TargetType); err != nil {
		return invalid("target_type", "invalid")
	}
	if _, err := localeSchema.Parse(in.Locale); err != nil {
		return invalid("locale", "invalid")
	}
	if err := noteSchema.Check(in.Note); err != nil {
		return invalid("note", "invalid")
	}
	if in.Kind == "merge" && in.OtherType != in.TargetType {
		return invalid("other_type", "type mismatch")
	}
	if len(in.SourceURLs) > 3 {
		return invalid("source_urls", "too many")
	}
	for _, raw := range in.SourceURLs {
		if _, err := sourceURLSchema.Parse(raw); err != nil {
			return invalid("source_urls", "invalid")
		}
	}
	if in.Kind == "create" {
		if _, ok := in.Changes["title"]; !ok {
			return invalid("title", "required")
		}
	}
	if len(in.Changes) == 0 {
		return invalid("changes", "empty")
	}

	needsSource := false
	effective := 0
	for field, change := range in.Changes {
		eff, src, err := validateField(in, field, change, now)
		if err != nil {
			return err
		}
		if eff {
			effective++
		}
		if src {
			needsSource = true
		}
	}
	if effective == 0 {
		return invalid("changes", "empty")
	}
	if needsSource && len(in.SourceURLs) == 0 {
		return invalid("source_urls", "required")
	}

	return nil
}

func validateField(in SuggestionInput, field string, change FieldChange, now time.Time) (bool, bool, error) {
	group, locale, provider, err := classify(field)
	if err != nil {
		return false, false, err
	}
	if locale != "" {
		if _, err := localeSchema.Parse(locale); err != nil {
			return false, false, invalid(field, "locale")
		}
	}
	if group == "merge" && in.Kind != "merge" {
		return false, false, invalid(field, "unknown")
	}

	switch group {
	case "title":
		eff, err := textChange(field, change, titleSchema)
		return eff, true, err
	case "synopsis":
		eff, err := textChange(field, change, synopsisSchema)
		return eff, true, err
	case "date":
		eff, err := releaseDateChange(field, change, now)
		return eff, true, err
	case "external":
		eff, err := textChange(field, change, externalSchema(provider))
		return eff, true, err
	case "cover":
		eff, err := textChange(field, change, coverSchema)
		return eff, false, err
	case "merge":
		eff, err := textChange(field, change, idSchema)
		return eff, false, err
	case "create":
		if in.Kind != "create" {
			return false, false, invalid(field, "unknown")
		}
		schema := map[string]z.StringSchema{
			"entity_kind": entityKindSchema, "work_id": uuidSchema, "language": languageSchema,
		}[field]
		eff, err := textChange(field, change, schema)
		return eff, false, err
	case "relation":
		if err := validateRelation(field, change, in.TargetType); err != nil {
			return false, false, err
		}
		n := len(change.Add) + len(change.Remove)
		src := field == "creators" || field == "organizations"
		return n > 0, src && n > 0, nil
	default:
		return false, false, invalid(field, "unknown")
	}
}

func classify(field string) (string, string, string, error) {
	switch field {
	case "title":
		return "title", "", "", nil
	case "release_date":
		return "date", "", "", nil
	case "cover":
		return "cover", "", "", nil
	case "*merge*":
		return "merge", "", "", nil
	case "entity_kind", "work_id", "language":
		return "create", "", "", nil
	case "genres", "creators", "organizations", "franchise", "platforms":
		return "relation", "", "", nil
	}
	if rest, ok := strings.CutPrefix(field, "translations."); ok {
		locale, rest, ok := strings.Cut(rest, ".")
		if ok && rest == "title" && locale != "" {
			return "title", locale, "", nil
		}
	}
	if locale, ok := strings.CutPrefix(field, "synopsis."); ok && locale != "" && !strings.Contains(locale, ".") {
		return "synopsis", locale, "", nil
	}
	if provider, ok := strings.CutPrefix(field, "external_ids."); ok {
		switch provider {
		case "wikidata", "tmdb", "igdb", "openlibrary", "isbn13":
			return "external", "", provider, nil
		}
	}

	return "", "", "", invalid(field, "unknown")
}

func externalSchema(provider string) z.StringSchema {
	switch provider {
	case "wikidata":
		return wikidataSchema
	case "tmdb", "igdb":
		return digitsSchema
	case "openlibrary":
		return openLibrarySchema
	default:
		return z.String().Refine(func(s string) error {
			if !catalog.ValidISBN13(s) {
				return errors.New("isbn")
			}

			return nil
		})
	}
}

func textChange(field string, change FieldChange, schema z.StringSchema) (bool, error) {
	after, afterNull, err := decodeString(change.After)
	if err != nil || afterNull {
		return false, invalid(field, "after")
	}
	normal, err := schema.Parse(after)
	if err != nil {
		return false, invalid(field, "invalid")
	}
	before, beforeNull, err := decodeString(change.Before)
	if err != nil {
		return false, invalid(field, "before")
	}
	if beforeNull {
		return true, nil
	}

	return norm.NFC.String(before) != norm.NFC.String(normal), nil
}

func releaseDateChange(field string, change FieldChange, now time.Time) (bool, error) {
	if err := validateReleaseDate(change.After, now); err != nil {
		return false, invalid(field, "invalid")
	}
	if bytes.Equal(bytes.TrimSpace(change.Before), []byte("null")) || len(bytes.TrimSpace(change.Before)) == 0 {
		return true, nil
	}
	if err := validateReleaseDate(change.Before, now); err != nil {
		return false, invalid(field, "before")
	}

	return canonical(change.Before) != canonical(change.After), nil
}

func validateReleaseDate(raw json.RawMessage, now time.Time) error {
	var dated struct {
		Value     string `json:"value"`
		Precision string `json:"precision"`
	}
	if err := json.Unmarshal(raw, &dated); err != nil {
		return err
	}
	if _, err := precisionSchema.Parse(dated.Precision); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", dated.Value)
	if err != nil || t.Format("2006-01-02") != dated.Value {
		return errors.New("date")
	}
	floor := time.Date(1870, 1, 1, 0, 0, 0, 0, time.UTC)
	ceil := dateOnly(now).AddDate(10, 0, 0)
	if t.Before(floor) || t.After(ceil) {
		return errors.New("range")
	}

	return nil
}

func validateRelation(field string, change FieldChange, targetType string) error {
	if len(change.Add)+len(change.Remove) > 20 {
		return invalid(field, "too many ops")
	}
	entityKind := ""
	switch field {
	case "creators":
		entityKind = "person"
	case "organizations":
		entityKind = "organization"
	}
	if err := checkOps(field, change.Add, entityKind, targetType); err != nil {
		return err
	}

	return checkOps(field, change.Remove, entityKind, targetType)
}

func checkOps(field string, ops []RelationOp, entityKind, targetType string) error {
	for _, op := range ops {
		if _, err := idSchema.Parse(op.ID); err != nil {
			return invalid(field, "id")
		}
		if entityKind == "" {
			continue
		}
		if !catalog.KnownRole(op.Role) || !catalog.RoleCompatible(op.Role, entityKind, targetType) {
			return invalid(field, "role")
		}
	}

	return nil
}

func decodeString(raw json.RawMessage) (string, bool, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", true, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false, err
	}

	return s, false, nil
}

func canonical(raw json.RawMessage) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}

	return string(b)
}

func dateOnly(t time.Time) time.Time {
	t = t.UTC()

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func rejectLineControls(s string) error {
	if strings.ContainsFunc(s, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\t' || unicode.IsControl(r)
	}) {
		return errors.New("control character")
	}

	return nil
}

func allowedSourceURL(s string) error {
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return errors.New("https required")
	}
	host := strings.ToLower(u.Hostname())
	if host == "hexascore.app" || strings.HasSuffix(host, ".hexascore.app") {
		return errors.New("hexascore.app refused")
	}
	if isShortener(host) {
		return errors.New("shortener refused")
	}

	return nil
}

func allowedCover(s string) error {
	if !strings.Contains(s, "://") {
		return nil
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return errors.New("https required")
	}

	return nil
}

func isShortener(host string) bool {
	if _, ok := shortenerHosts[host]; ok {
		return true
	}
	for name := range shortenerHosts {
		if strings.HasSuffix(host, "."+name) {
			return true
		}
	}

	return false
}
