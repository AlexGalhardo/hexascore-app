package models

import (
	"context"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/orm"
)

// User is a registered account (vukem install:auth).
type User struct {
	orm.Model

	Name            string       `json:"name"`
	Email           string       `json:"email"`
	Password        string       `json:"-"`
	EmailVerifiedAt orm.NullTime `json:"email_verified_at"`
	SessionVersion  int64        `json:"-"`
	// Role is "user", "moderator" or "admin"; it is never writable from a form.
	Role         string       `json:"role"`
	BannedAt     orm.NullTime `json:"banned_at"`
	BanReason    *string      `json:"ban_reason"`
	BanExpiresAt orm.NullTime `json:"ban_expires_at"`
	// DeletionScheduledAt starts the 30-day grace period; AnonymizedAt marks a purged account.
	DeletionScheduledAt orm.NullTime `json:"deletion_scheduled_at"`
	AnonymizedAt        orm.NullTime `json:"-"`
	// IsSeed marks a fake legacy seed user: it can never sign in and `seed:purge` deletes it.
	IsSeed bool `json:"-"`
}

// RoleAdmin is the Role of administrators.
const RoleAdmin = "admin"

// RoleModerator may moderate community suggestions while the community_trusted_moderators flag is on.
const RoleModerator = "moderator"

// BeforeCreate gives new accounts the default role (the column has a CHECK).
func (u *User) BeforeCreate(context.Context) error {
	if u.Role == "" {
		u.Role = "user"
	}

	return nil
}

// IsBanned reports whether the ban is active at now: a ban whose expiry has
// passed no longer counts (SDD 03 section 7.1).
func (u User) IsBanned(now time.Time) bool {
	return u.BannedAt.Valid && (!u.BanExpiresAt.Valid || u.BanExpiresAt.V.After(now))
}

// AuthID implements auth.Authenticatable.
func (u User) AuthID() string { return strconv.FormatInt(u.ID, 10) }

// AuthPassword implements auth.Authenticatable.
func (u User) AuthPassword() string { return u.Password }

// AuthSessionVersion implements auth.SessionVersioner: sessions logged in
// with another version are signed out.
func (u User) AuthSessionVersion() int64 { return u.SessionVersion }

// EmailVerified implements auth.Verifiable.
func (u User) EmailVerified() bool { return u.EmailVerifiedAt.Valid }
