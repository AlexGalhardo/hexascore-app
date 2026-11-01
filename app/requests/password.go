package requests

import (
	"regexp"

	"github.com/AlexGalhardo/vukem/z"
)

// Password length bounds in runes. Owner round 2026-11-01 item 8 sets the
// maximum back to the legacy 32 (it was 128 under D-24); logins never check it,
// so an older, longer password still signs in.
const (
	PasswordMin = 8
	PasswordMax = 32
)

// PasswordSchema is the rule set for a new password (sign-up, reset, change,
// set), never for login. It mirrors legacy lib/validation.ts: upper, lower,
// digit and a non-alphanumeric character (a space counts, as in legacy).
var PasswordSchema = z.String().Min(PasswordMin).Max(PasswordMax).
	HasUpper().HasLower().HasDigit().
	Regex(regexp.MustCompile(`[^A-Za-z0-9]`))
