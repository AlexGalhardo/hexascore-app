package legacyimport

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
)

// check is one pre-flight assert of SDD 03 §5: a count that must be zero.
type check struct {
	name, query string
}

// preflightChecks run against the restored legacy schema.
var preflightChecks = []check{
	{"duplicate_emails", `SELECT count(*) FROM (SELECT lower(email) FROM legacy_import."user"
GROUP BY 1 HAVING count(*) > 1) d`},
	{"twofactor_without_row", `SELECT count(*) FROM legacy_import."user" u
WHERE u.two_factor_enabled AND NOT EXISTS (SELECT 1 FROM legacy_import.two_factor t WHERE t.user_id = u.id)`},
	// No password, no Google link and no passkey: nothing they could sign in with.
	{"no_login_method", `SELECT count(*) FROM legacy_import."user" u
WHERE NOT EXISTS (SELECT 1 FROM legacy_import.account a WHERE a.user_id = u.id
  AND ((a.provider_id = 'credential' AND a.password IS NOT NULL) OR a.provider_id = 'google'))
  AND NOT EXISTS (SELECT 1 FROM legacy_import.passkey p WHERE p.user_id = u.id)`},
}

// passkeyOnlySQL counts users whose only login is a passkey: not an abort,
// they get the reset notice (SDD 03 §5, P2.F.2.2).
const passkeyOnlySQL = `SELECT count(*) FROM legacy_import."user" u
WHERE NOT EXISTS (SELECT 1 FROM legacy_import.account a WHERE a.user_id = u.id
  AND ((a.provider_id = 'credential' AND a.password IS NOT NULL) OR a.provider_id = 'google'))
  AND EXISTS (SELECT 1 FROM legacy_import.passkey p WHERE p.user_id = u.id)`

// PreflightResult holds the assert counts.
type PreflightResult struct {
	Failures    map[string]int64
	PasskeyOnly int64
}

// PreflightError aborts the import; it names every failing assert.
type PreflightError struct{ Failures map[string]int64 }

func (e *PreflightError) Error() string {
	var parts []string

	for _, c := range preflightChecks {
		if n := e.Failures[c.name]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", c.name, n))
		}
	}

	return "legacyimport: pre-flight failed: " + strings.Join(parts, " ")
}

// Preflight runs the asserts; a non-zero count returns *PreflightError
// with the result, which still carries every count.
func Preflight(ctx context.Context, db *database.DB) (PreflightResult, error) {
	res := PreflightResult{Failures: map[string]int64{}}

	for _, c := range preflightChecks {
		var n int64
		if err := db.QueryRowContext(ctx, c.query).Scan(&n); err != nil {
			return res, fmt.Errorf("legacyimport: pre-flight %s: %w", c.name, err)
		}

		if n > 0 {
			res.Failures[c.name] = n
		}
	}

	if err := db.QueryRowContext(ctx, passkeyOnlySQL).Scan(&res.PasskeyOnly); err != nil {
		return res, fmt.Errorf("legacyimport: pre-flight passkey_only: %w", err)
	}

	if len(res.Failures) > 0 {
		return res, &PreflightError{Failures: res.Failures}
	}

	return res, nil
}
