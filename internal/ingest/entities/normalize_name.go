// Package entities resolves the credits of an item to entities rows (SDD 06 §7.4 "Matching per credit").
package entities

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// folds are the Latin letters NFKD does not decompose; they match what unaccent() (hs_norm in SQL) does.
var folds = strings.NewReplacer("ø", "o", "ł", "l", "đ", "d", "ð", "d", "ß", "ss", "æ", "ae", "œ", "oe", "þ", "th")

// NormalizeName is the name key of step 3: NFKD, combining marks dropped, Latin folded to ASCII, lowercase and
// every run of punctuation or space as one space. Non-Latin letters stay, so a CJK name still matches itself.
// It mirrors the SQL function hs_norm, which indexes entities.default_name.
func NormalizeName(s string) string {
	var b strings.Builder

	for _, r := range norm.NFKD.String(strings.ToLower(s)) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}

	return strings.Join(strings.Fields(folds.Replace(b.String())), " ")
}
