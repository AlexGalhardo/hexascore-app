package legacyimport

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/auth/hash"
	"github.com/AlexGalhardo/vukem/database"
)

// legacyUser is one Better Auth user with its credential hash, if any.
type legacyUser struct {
	id, name, email                 string
	emailVerified, banned, private  bool
	createdAt                       time.Time
	updatedAt, banExpires, deletion sql.NullTime
	image, username, display, role  sql.NullString
	banReason, password             sql.NullString
}

// usersBatch maps Better Auth users onto users (SDD 03 §5 mapping table) and
// their Google accounts onto social_accounts. Sessions, verifications, rate
// limits and OAuth tokens are not read.
func usersBatch(hasher hash.Hasher, now func() time.Time) BatchFunc {
	return func(ctx context.Context, tx *database.Tx, after string, n int) (string, int, error) {
		users, err := readUsers(ctx, tx, after, n)
		if err != nil || len(users) == 0 {
			return "", 0, err
		}

		at := now().UTC()

		for _, u := range users {
			if err := upsertUser(ctx, tx, hasher, u, at); err != nil {
				return "", 0, fmt.Errorf("user %s: %w", u.id, err)
			}
		}

		last := users[len(users)-1].id

		// Google links keep (provider, provider_id) so sign-in survives the cutover (§7.7).
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO social_accounts (user_id, provider, provider_id, created_at, updated_at)
SELECT n.id, 'google', a.account_id, a.created_at, $3
FROM legacy_import.account a JOIN users n ON n.legacy_id = a.user_id
WHERE a.provider_id = 'google' AND a.user_id > $1 AND a.user_id <= $2
ON CONFLICT (provider, provider_id) DO NOTHING`,
			after,
			last,
			at,
		); err != nil {
			return "", 0, fmt.Errorf("google accounts: %w", err)
		}

		return last, len(users), nil
	}
}

func readUsers(ctx context.Context, tx *database.Tx, after string, n int) ([]legacyUser, error) {
	rows, err := tx.QueryContext(
		ctx,
		`SELECT u.id, u.name, u.email, u.email_verified, u.image, u.created_at, u.updated_at,
  u.username, u.display_username, u.role, coalesce(u.banned, false), u.ban_reason, u.ban_expires, u.profile_private,
  u.deletion_scheduled_at,
  (SELECT a.password FROM legacy_import.account a WHERE a.user_id = u.id AND a.provider_id = 'credential'
     AND a.password IS NOT NULL ORDER BY a.id LIMIT 1)
FROM legacy_import."user" u WHERE u.id > $1 ORDER BY u.id LIMIT $2`,
		after,
		n,
	)
	if err != nil {
		return nil, fmt.Errorf("read users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []legacyUser

	for rows.Next() {
		var u legacyUser
		if err := rows.Scan(&u.id, &u.name, &u.email, &u.emailVerified, &u.image, &u.createdAt, &u.updatedAt,
			&u.username, &u.display, &u.role, &u.banned, &u.banReason, &u.banExpires, &u.private, &u.deletion,
			&u.password); err != nil {
			return nil, fmt.Errorf("read users: %w", err)
		}

		out = append(out, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read users: %w", err)
	}

	return out, nil
}

func upsertUser(ctx context.Context, tx *database.Tx, hasher hash.Hasher, u legacyUser, at time.Time) error {
	// The Better Auth hash is copied verbatim; CompatHasher verifies and rehashes it on login (§7.2).
	password, passwordSetAt := u.password.String, sql.NullTime{Time: at, Valid: true}
	if !u.password.Valid {
		// Google- or passkey-only: a password nobody knows, until the user sets one.
		h, err := hasher.Make(rand.Text())
		if err != nil {
			return fmt.Errorf("random password: %w", err)
		}

		password, passwordSetAt = h, sql.NullTime{}
	}

	var verifiedAt sql.NullTime
	if u.emailVerified {
		verifiedAt = sql.NullTime{Time: u.createdAt, Valid: true}
		if u.updatedAt.Valid {
			verifiedAt.Time = u.updatedAt.Time
		}
	}

	role := "user"
	if slices.ContainsFunc(
		strings.Split(u.role.String, ","),
		func(r string) bool { return strings.TrimSpace(r) == "admin" },
	) {
		role = "admin"
	}

	// Expired bans are dropped; a live ban starts at migration time.
	var bannedAt, banExpires sql.NullTime

	banReason := sql.NullString{}
	if u.banned && (!u.banExpires.Valid || u.banExpires.Time.After(at)) {
		bannedAt, banExpires, banReason = sql.NullTime{Time: at, Valid: true}, u.banExpires, u.banReason
	}

	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO users (legacy_id, name, email, email_verified_at, password, password_set_at,
  username, display_username, image_url, profile_private, role, banned_at, ban_reason, ban_expires_at,
  deletion_scheduled_at, created_at, updated_at)
VALUES ($1, $2, lower($3), $4, $5, $6, lower($7), $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
ON CONFLICT (legacy_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email,
  email_verified_at = EXCLUDED.email_verified_at, password = EXCLUDED.password, password_set_at = EXCLUDED.password_set_at,
  username = EXCLUDED.username, display_username = EXCLUDED.display_username, image_url = EXCLUDED.image_url,
  profile_private = EXCLUDED.profile_private, role = EXCLUDED.role, banned_at = EXCLUDED.banned_at,
  ban_reason = EXCLUDED.ban_reason, ban_expires_at = EXCLUDED.ban_expires_at,
  deletion_scheduled_at = EXCLUDED.deletion_scheduled_at, updated_at = EXCLUDED.updated_at`,
		u.id,
		u.name,
		u.email,
		verifiedAt,
		password,
		passwordSetAt,
		u.username,
		u.display,
		u.image,
		u.private,
		role,
		bannedAt,
		banReason,
		banExpires,
		u.deletion,
		u.createdAt,
		at,
	)

	return err
}
