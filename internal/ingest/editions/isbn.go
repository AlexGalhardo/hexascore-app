// Package editions turns the book editions found for a work into editions rows (SDD 06 §7.4, ADR 0012).
package editions

import (
	"strconv"
	"strings"
)

// ToISBN13 strips raw to digits and X, then returns the ISBN-13 it stands for: a 10-character ISBN with a
// correct mod-11 check is converted (978 + its first 9 digits + the new check digit), a 13-digit one needs the
// 978/979 prefix and a correct check digit. Anything else is rejected.
func ToISBN13(raw string) (string, bool) {
	var b strings.Builder

	for _, r := range strings.ToUpper(raw) {
		if r >= '0' && r <= '9' || r == 'X' {
			b.WriteRune(r)
		}
	}

	s := b.String()

	switch len(s) {
	case 10:
		if !validISBN10(s) {
			return "", false
		}

		core := "978" + s[:9]

		return core + strconv.Itoa(check13(core)), true
	case 13:
		if strings.Contains(s, "X") || s[:3] != "978" && s[:3] != "979" || check13(s[:12]) != int(s[12]-'0') {
			return "", false
		}

		return s, true
	}

	return "", false
}

func validISBN10(s string) bool {
	sum := 0

	for i := range 10 {
		d := 10

		if s[i] != 'X' {
			d = int(s[i] - '0')
		} else if i != 9 {
			return false
		}

		sum += d * (10 - i)
	}

	return sum%11 == 0
}

// check13 is the EAN-13 check digit of the first 12 digits.
func check13(s string) int {
	sum := 0

	for i := range 12 {
		d := int(s[i] - '0')
		if i%2 == 1 {
			d *= 3
		}

		sum += d
	}

	return (10 - sum%10) % 10
}
