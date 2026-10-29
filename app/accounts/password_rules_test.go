package accounts_test

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
)

func TestPasswordRules(t *testing.T) {
	t.Parallel()

	pad := func(n int) string { return "Aa1!" + strings.Repeat("x", n-4) }

	for _, tc := range []struct {
		name, password string
		ok             bool
	}{
		{"128 characters accepted", pad(128), true},
		{"129 characters rejected", pad(129), false},
		{"8 characters accepted", "Abcde1!x", true},
		{"7 characters rejected", "Abcd1!x", false},
		{"128 multibyte runes accepted", "Aa1!" + strings.Repeat("é", 124), true},
		{"no uppercase", "abcde1!x", false},
		{"no lowercase", "ABCDE1!X", false},
		{"no digit", "Abcdef!x", false},
		{"no special", "Abcdefg1", false},
		{"space counts as special", "Abcdef 1", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := accounts.ValidatePassword(tc.password); (err == nil) != tc.ok {
				t.Fatalf("ValidatePassword err = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}
