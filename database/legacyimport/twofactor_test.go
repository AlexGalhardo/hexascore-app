//go:build integration

package legacyimport_test

import (
	"database/sql"
	"encoding/base32"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/auth/totp"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

// betterAuthSecret has the shape Better Auth 1.7.7 generates: 32 characters
// of a-z, A-Z and 0-9, used as raw HMAC key bytes.
const betterAuthSecret = "q8Zr1LmN4pXv7TbW2yHc9KdF3sJg6EaU"

func TestTwoFactorUsersPassChallenge(t *testing.T) {
	t.Parallel()

	app, db := fixtureDB(t)
	ctx := t.Context()

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"users", "twofactor"})
	if err != nil {
		t.Fatal(err)
	}

	export := `{"userId":"u02","secret":"` + betterAuthSecret + `","backupCodes":["aB3dE-fG4hI","Zx9Yw-Vu8Ts"]}
{"userId":"u04","secret":"` + betterAuthSecret + `","backupCodes":[]}
`
	stages[1].Batch = legacyimport.TwoFactorBatch(strings.NewReader(export), time.Now)

	r := &legacyimport.Runner{DB: db, Batch: 1, Out: io.Discard, Now: time.Now}
	if err := r.Run(ctx, stages); err != nil {
		t.Fatalf("run: %v", err)
	}

	var confirmed sql.NullTime

	var stored sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT two_factor_confirmed_at, two_factor_secret FROM users WHERE legacy_id = 'u04'`).
		Scan(&confirmed, &stored); err != nil {
		t.Fatal(err)
	}

	if confirmed.Valid || !stored.Valid || strings.Contains(stored.String, betterAuthSecret) {
		t.Fatalf(
			"u04: confirmed=%v sealed=%v: an unverified row stays pending, and the secret is sealed",
			confirmed.Valid,
			stored.Valid,
		)
	}

	// A code from the legacy secret passes the challenge after the password step.
	code, err := totp.Code(base32.StdEncoding.EncodeToString([]byte(betterAuthSecret)), time.Now(), totp.Options{})
	if err != nil {
		t.Fatal(err)
	}

	login := func() *vukemtest.Client {
		c := app.Client(t).Inertia("").WithHeader("Referer", "/login")
		c.PostJSON("/login", map[string]any{"identifier": "bruno@example.test", "password": "correct horse battery staple"}).
			AssertRedirect("/two-factor-challenge")

		return c
	}

	login().PostJSON("/two-factor-challenge", map[string]any{"code": code, "recovery_code": ""}).
		AssertRedirect("/onboarding")

	// An unused Better Auth backup code works once as a recovery code.
	login().PostJSON("/two-factor-challenge", map[string]any{"code": "", "recovery_code": "aB3dE-fG4hI"}).
		AssertRedirect("/onboarding")
	login().WithHeader("Referer", "/two-factor-challenge").
		PostJSON("/two-factor-challenge", map[string]any{"code": "", "recovery_code": "aB3dE-fG4hI"}).
		AssertRedirect("/two-factor-challenge")
}
