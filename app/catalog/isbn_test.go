package catalog

import "testing"

func TestISBN13(t *testing.T) {
	cases := map[string]bool{
		"9780306406157":  true,
		"9780306406158":  false,
		"9791032305690":  true,
		"9770306406156":  false, // 977 prefix is not a book
		"978030640615":   false,
		"97803064061570": false,
		"97803064O6157":  false,
		"":               false,
	}
	for isbn, want := range cases {
		if got := ValidISBN13(isbn); got != want {
			t.Errorf("ValidISBN13(%q) = %v, want %v", isbn, got, want)
		}
	}
}
