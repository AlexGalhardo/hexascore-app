//go:build integration

package tests

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type failingCanceller struct{ ids map[int64]bool }

func (f failingCanceller) CancelLive(_ context.Context, id int64) error {
	if f.ids[id] {
		return errors.New("stripe down")
	}

	return nil
}

func deletionApp(t *testing.T) (*vukemtest.App, *accounts.Accounts, *database.DB) {
	t.Helper()

	opts := vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}}
	app := vukemtest.New(t, bootstrap.New, opts)

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	return app, acc, db
}

func count(t *testing.T, db *database.DB, q string, args ...any) int {
	t.Helper()

	var n int
	if err := db.QueryRowContext(t.Context(), q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func TestDeletionScheduleAndCancel(t *testing.T) {
	app, _, db := deletionApp(t)

	c := app.Client(t)
	register(t, app, c, "ana@example.com")

	app.Client(t).
		PostJSON("/profile/deletion", map[string]any{"confirm": true}).
		AssertStatus(401)
		// guests cannot schedule
	c.PostJSON("/profile/deletion", map[string]any{"confirm": false}).AssertUnprocessable()
	c.PostJSON("/profile/deletion", map[string]any{"confirm": true}).AssertRedirect("/login")
	c.Get("/dashboard").AssertRedirect("/login")

	if n := count(t, db, `SELECT count(*) FROM users WHERE deletion_scheduled_at IS NOT NULL`); n != 1 {
		t.Fatalf("scheduled users = %d", n)
	}

	if n := count(t, db, `SELECT count(*) FROM account_events WHERE kind = 'deletion_scheduled'`); n != 1 {
		t.Fatalf("deletion_scheduled events = %d", n)
	}

	// The grace period lets the user back in to cancel.
	c.PostJSON("/login", map[string]any{"identifier": "ana@example.com", "password": testPassword}).
		AssertRedirect("/onboarding") // verified, no nickname yet (P1.C.2.1b)
	c.Get("/dashboard").AssertOK()
	c.Delete("/profile/deletion").AssertRedirect("/dashboard")
	c.Delete("/profile/deletion").AssertStatus(409)

	if n := count(t, db, `SELECT count(*) FROM users WHERE deletion_scheduled_at IS NOT NULL`); n != 0 {
		t.Fatalf("still scheduled: %d", n)
	}

	app.Client(t).Get("/help/delete-account").AssertOK().AssertInertia("Help/DeleteAccount")
}

func TestDeletionPurgeAnonymisesKeepsRatings(t *testing.T) {
	app, acc, db := deletionApp(t)
	ctx := t.Context()

	ids := map[string]int64{"gone@example.com": 0, "stripe@example.com": 0, "fresh@example.com": 0}
	for e := range ids {
		register(t, app, app.Client(t), e)

		var id int64
		if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, e).Scan(&id); err != nil {
			t.Fatal(err)
		}

		ids[e] = id
	}

	gone, blocked, fresh := ids["gone@example.com"], ids["stripe@example.com"], ids["fresh@example.com"]

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://wikidata.org')`,
	); err != nil {
		t.Fatal(err)
	}

	var item string

	err := db.QueryRowContext(
		ctx,
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES (gen_random_uuid(), 'game', 'g', 'Q1') RETURNING id::text`,
	).
		Scan(&item)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6)
VALUES ($1, $3, 50, 50, 50, 50, 50, 50), ($2, $3, 60, 60, 60, 60, 60, 60)`, gone, blocked, item); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO passkeys (user_id, credential_id, public_key, algorithm, name, created_at)
VALUES ($1, 'cred-gone', 'pk', -7, 'Passkey', now()), ($2, 'cred-blocked', 'pk', -7, 'Passkey', now())`,
		gone,
		blocked,
	); err != nil {
		t.Fatal(err)
	}

	goneID := strconv.FormatInt(gone, 10)
	if _, _, err := acc.Tokens.Create(ctx, goneID, "cli", nil, time.Hour); err != nil {
		t.Fatal(err)
	}

	old := time.Now().UTC().AddDate(0, 0, -31)
	for i, id := range []int64{gone, blocked} {
		nick := "nick" + strings.Repeat("a", i)
		if _, err := db.ExecContext(
			ctx,
			`UPDATE users SET deletion_scheduled_at = $1, username = $3 WHERE id = $2`,
			old,
			id,
			nick,
		); err != nil {
			t.Fatal(err)
		}
	}

	acc.Subscriptions = failingCanceller{ids: map[int64]bool{blocked: true}}

	done, err := acc.PurgeDue(ctx)
	if err != nil || len(done) != 1 || done[0] != gone {
		t.Fatalf("purged %v, %v; want only %d", done, err, gone)
	}

	if n := count(t, db, `SELECT count(*) FROM ratings`); n != 2 {
		t.Fatalf("ratings = %d, want 2", n)
	}

	var email, name string

	var anonymised bool

	if err := db.QueryRowContext(ctx, `SELECT email, name, anonymized_at IS NOT NULL FROM users WHERE id = $1`, gone).
		Scan(&email, &name, &anonymised); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(email, "deleted-") || !strings.HasSuffix(email, "@invalid") || name != "Deleted user" ||
		!anonymised {
		t.Fatalf("not anonymised: %q %q %v", email, name, anonymised)
	}

	cleared := `SELECT count(*) FROM users WHERE id = $1 AND username IS NULL AND two_factor_secret IS NULL
AND profile_private AND deletion_scheduled_at IS NULL AND email_verified_at IS NULL`
	if n := count(t, db, cleared, gone); n != 1 {
		t.Fatal("PII columns not cleared")
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM access_tokens WHERE user_id = $1 AND revoked_at IS NULL`,
		goneID,
	); n != 0 {
		t.Fatalf("live tokens left: %d", n)
	}

	// The user row survives anonymisation, so ON DELETE CASCADE never removes its passkeys.
	if n := count(t, db, `SELECT count(*) FROM passkeys WHERE user_id = $1`, gone); n != 0 {
		t.Fatalf("passkeys left after purge: %d", n)
	}

	if n := count(t, db, `SELECT count(*) FROM passkeys WHERE user_id = $1`, blocked); n != 1 {
		t.Fatal("blocked user's passkey was touched")
	}

	// A Stripe failure keeps the user scheduled and intact.
	intact := `SELECT count(*) FROM users WHERE id = $1 AND deletion_scheduled_at IS NOT NULL
AND anonymized_at IS NULL AND email = 'stripe@example.com'`
	if n := count(t, db, intact, blocked); n != 1 {
		t.Fatal("blocked user was touched")
	}

	if n := count(t, db, `SELECT count(*) FROM users WHERE id = $1 AND anonymized_at IS NULL`, fresh); n != 1 {
		t.Fatal("unscheduled user was touched")
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM account_events WHERE user_id = $1 AND kind = 'anonymized'`,
		gone,
	); n != 1 {
		t.Fatal("missing anonymized event")
	}

	// Idempotent, and a recovered Stripe lets the retry through.
	acc.Subscriptions = failingCanceller{}
	if done, err := acc.PurgeDue(ctx); err != nil || len(done) != 1 || done[0] != blocked {
		t.Fatalf("retry purged %v, %v", done, err)
	}

	if done, _ := acc.PurgeDue(ctx); len(done) != 0 {
		t.Fatalf("third run purged %v", done)
	}

	app.Client(t).PostJSON("/login", map[string]any{"identifier": "gone@example.com", "password": testPassword}).
		AssertStatus(422)
}
