//go:build integration

package legacyimport_test

import (
	"database/sql"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

type importedUser struct {
	email, password, role                                string
	verifiedAt, passwordSetAt, bannedAt, deletion, banEx sql.NullTime
	username                                             sql.NullString
	private                                              bool
}

func TestUsersStage(t *testing.T) {
	t.Parallel()

	app, db := fixtureDB(t)
	ctx := t.Context()

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"users"})
	if err != nil {
		t.Fatal(err)
	}

	r := &legacyimport.Runner{DB: db, Batch: 3, Out: io.Discard, Now: time.Now}
	if err := r.Run(ctx, stages); err != nil {
		t.Fatalf("run: %v", err)
	}

	got := map[string]importedUser{}

	rows, err := db.QueryContext(ctx, `SELECT legacy_id, email, password, role, email_verified_at, password_set_at,
  banned_at, deletion_scheduled_at, ban_expires_at, username, profile_private FROM users WHERE legacy_id IS NOT NULL`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			id string
			u  importedUser
		)
		if err := rows.Scan(&id, &u.email, &u.password, &u.role, &u.verifiedAt, &u.passwordSetAt, &u.bannedAt,
			&u.deletion, &u.banEx, &u.username, &u.private); err != nil {
			t.Fatal(err)
		}

		got[id] = u
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if len(got) != 7 {
		t.Fatalf("imported %d users, want 7", len(got))
	}

	ana := got["u01"]
	if ana.email != "ana@example.test" || !ana.verifiedAt.Valid || !ana.verifiedAt.Time.Equal(
		time.Date(2025, 1, 3, 11, 0, 0, 0, time.UTC),
	) || ana.username.String != "anaana" || !ana.private {
		t.Errorf(
			"u01: email=%s verified=%v username=%v private=%v, want lowercased email, verified at legacy updated_at",
			ana.email,
			ana.verifiedAt,
			ana.username,
			ana.private,
		)
	}

	if !strings.HasPrefix(ana.password, "f784aeba634fcc3813d2fed8886d1f4f:") || !ana.passwordSetAt.Valid {
		t.Error("u01: credential hash not copied verbatim or password_set_at missing")
	}

	if c := got["u03"]; c.verifiedAt.Valid || !c.deletion.Valid ||
		!c.deletion.Time.Equal(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("u03: verified=%v deletion=%v, want unverified and deletion grace preserved", c.verifiedAt.Valid,
			c.deletion.Time.UTC())
	}

	if d := got["u04"]; d.role != "admin" || !d.bannedAt.Valid || !d.banEx.Valid {
		t.Errorf(
			"u04: role=%s banned=%v expires=%v, want admin with a live ban",
			d.role,
			d.bannedAt.Valid,
			d.banEx.Valid,
		)
	}

	if f := got["u06"]; f.bannedAt.Valid || f.banEx.Valid || f.role != "user" {
		t.Errorf("u06: role=%s banned=%v, expired ban must be dropped", f.role, f.bannedAt.Valid)
	}

	for _, id := range []string{"u03", "u05", "u07"} {
		if u := got[id]; !strings.HasPrefix(u.password, "$argon2id$") || u.passwordSetAt.Valid {
			t.Errorf("%s: want random argon2id with password_set_at NULL, set=%v", id, u.passwordSetAt.Valid)
		}
	}

	if got["u03"].password == got["u05"].password {
		t.Error("random passwords must differ per user")
	}

	var links string
	if err := db.QueryRowContext(ctx, `SELECT string_agg(u.legacy_id || '=' || s.provider || ':' || s.provider_id, ','
  ORDER BY u.legacy_id) FROM social_accounts s JOIN users u ON u.id = s.user_id`).Scan(&links); err != nil {
		t.Fatal(err)
	}

	if links != "u03=google:g-1003,u05=google:g-1005,u06=google:g-1006" {
		t.Errorf("social accounts = %s", links)
	}

	// An imported Better Auth scrypt hash signs in and is rehashed to argon2id (SDD 03 §7.2).
	c := app.Client(t).Inertia("").WithHeader("Referer", "/login")
	c.PostJSON("/login", map[string]any{"identifier": "ana@example.test", "password": "correct horse battery staple"})

	var stored string
	if err := db.QueryRowContext(ctx, `SELECT password FROM users WHERE legacy_id = 'u01'`).Scan(&stored); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(stored, "$argon2id$") {
		t.Fatalf("u01 hash after login = %.12q..., want $argon2id$", stored)
	}
}
