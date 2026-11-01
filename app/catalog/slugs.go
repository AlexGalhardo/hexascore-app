package catalog

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/AlexGalhardo/vukem/database"
	"golang.org/x/text/unicode/norm"
)

// MaxSlugLength matches the slug CHECK constraints (SDD 01 §5.2).
const MaxSlugLength = 120

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// reserved slugs collide with listing routes under the same section (SDD 01 §7.1), including the entity paths
// of EntityPaths: /games/platforms/{slug} would shadow the changes page of an item slugged "platforms".
var reserved = map[string]bool{
	"genres": true, "trending": true, "search": true, "new": true, "people": true, "companies": true,
	"franchises": true, "platforms": true, "game-modes": true, "countries": true, "languages": true,
}

// ErrEmptySlug means the title has no letter or digit to build a slug from.
var ErrEmptySlug = errors.New("catalog: title yields an empty slug")

// Slugify builds a slug from the en title: NFKD, marks stripped, lowercase,
// runs of non-alphanumerics become one "-", trimmed to MaxSlugLength.
func Slugify(title string) string {
	noMarks := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}

		return r
	}, norm.NFKD.String(title))

	slug := strings.Trim(nonAlnum.ReplaceAllString(strings.ToLower(noMarks), "-"), "-")
	if len(slug) > MaxSlugLength {
		slug = strings.TrimRight(slug[:MaxSlugLength], "-")
	}

	return slug
}

// UniqueSlug returns a free slug for title in the namespace (item type or
// entity kind): the plain slug, then with the release year (0 = unknown),
// then -2, -3, ... A slug is taken by a live row, by a redirect old_slug
// (invariant I-1) or by a reserved word. Call it in the transaction that
// inserts the row; the unique index still decides races.
func UniqueSlug(
	ctx context.Context,
	ex database.Executor,
	scope Scope,
	namespace, title string,
	year int,
) (string, error) {
	base := Slugify(title)
	if base == "" {
		return "", ErrEmptySlug
	}

	candidates := []string{base}
	if year > 0 {
		candidates = append(candidates, withSuffix(base, strconv.Itoa(year)))
	}

	for n := 2; ; n++ {
		candidates = append(candidates, withSuffix(base, strconv.Itoa(n)))

		slug := candidates[0]
		candidates = candidates[1:]

		if free, err := slugFree(ctx, ex, scope, namespace, slug); err != nil || free {
			return slug, err
		}
	}
}

// FirstFreeSlug is the search UniqueSlug runs without the year step, for callers that already know
// which slugs are taken: base, then -2, -3, ..., skipping reserved words.
func FirstFreeSlug(base string, taken func(slug string) bool) string {
	if !reserved[base] && !taken(base) {
		return base
	}

	for n := 2; ; n++ {
		if slug := withSuffix(base, strconv.Itoa(n)); !reserved[slug] && !taken(slug) {
			return slug
		}
	}
}

func withSuffix(base, suffix string) string {
	if room := MaxSlugLength - len(suffix) - 1; len(base) > room {
		base = strings.TrimRight(base[:room], "-")
	}

	return base + "-" + suffix
}

func slugFree(ctx context.Context, ex database.Executor, scope Scope, namespace, slug string) (bool, error) {
	if reserved[slug] {
		return false, nil
	}

	table, col, err := scope.table()
	if err != nil {
		return false, err
	}

	var taken bool

	q := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE %s = $1 AND slug = $2)
  OR EXISTS (SELECT 1 FROM catalog_redirects WHERE scope = $3 AND namespace = $1 AND old_slug = $2)`, table, col)
	if err := ex.QueryRowContext(ctx, q, namespace, slug, string(scope)).Scan(&taken); err != nil {
		return false, err
	}

	return !taken, nil
}
