package requests

import (
	"regexp"

	"github.com/AlexGalhardo/vukem/z"
)

// Password length bounds in runes. Legacy minimum 8; the maximum is raised
// from legacy 32 to 128 by D-24 (SDD 03 FR-03-2).
const (
	PasswordMin = 8
	PasswordMax = 128
)

// PasswordSchema is the rule set for a new password (sign-up, reset, change,
// set), never for login. It mirrors legacy lib/validation.ts: upper, lower,
// digit and a non-alphanumeric character (a space counts, as in legacy).
var PasswordSchema = z.String().Min(PasswordMin).Max(PasswordMax).
	HasUpper().HasLower().HasDigit().
	Regex(regexp.MustCompile(`[^A-Za-z0-9]`))
