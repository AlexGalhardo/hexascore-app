// Package normalize maps the stored source records of one Wikidata QID to
// the canonical catalogue rows (SDD 06 §7.3). The functions here are pure.
package normalize

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// MaxSlugLength is the ingest slug cap (legacy `slugify`).
const MaxSlugLength = 80

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify lowercases the title and keeps ASCII letters and digits only: NFKD,
// everything non-ASCII dropped (a CJK title yields ""), runs of other bytes
// become one "-", cut at MaxSlugLength without a trailing "-".
func Slugify(title string) string {
	ascii := strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return -1
		}

		return r
	}, norm.NFKD.String(title))

	slug := strings.Trim(nonAlnum.ReplaceAllString(strings.ToLower(ascii), "-"), "-")
	if len(slug) > MaxSlugLength {
		slug = strings.TrimRight(slug[:MaxSlugLength], "-")
	}

	return slug
}

// SlugCandidates lists the slugs to try in order: the title, then with the
// release year (0 = unknown), then with the lowercase QID. A title without
// ASCII content falls back to the QID alone. Choosing among them against the
// database, and never changing a stored slug, is the normalize job's work.
func SlugCandidates(title string, year int, qid string) []string {
	base := Slugify(title)
	if base == "" {
		return []string{strings.ToLower(qid)}
	}

	out := []string{base}
	if year != 0 {
		out = append(out, base+"-"+strconv.Itoa(year))
	}

	return append(out, base+"-"+strings.ToLower(qid))
}
