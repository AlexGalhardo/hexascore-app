package catalog

// ValidISBN13 mirrors the SQL function isbn13_is_valid: 13 digits, 978 or 979
// prefix, correct check digit.
func ValidISBN13(s string) bool {
	if len(s) != 13 || s[:2] != "97" || (s[2] != '8' && s[2] != '9') {
		return false
	}
	sum := 0
	for i := range 13 {
		d := int(s[i]) - '0'
		if d < 0 || d > 9 {
			return false
		}
		if i%2 == 1 {
			d *= 3
		}
		sum += d
	}
	return sum%10 == 0
}
