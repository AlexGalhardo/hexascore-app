package accounts

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/AlexGalhardo/vukem/auth"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const (
	// AppTokenTTL is the mobile token lifetime; there is no refresh token (SDD 03 §7.4).
	AppTokenTTL = 60 * 24 * time.Hour
	// MaxAppTokens is the live-token cap per user; the next issue revokes the least recently used.
	MaxAppTokens = 10
	// MaxDeviceName bounds the stored device name in runes.
	MaxDeviceName = 64
)

// AppAbilities are the six abilities every app token carries; never "*" (README "Access-token abilities").
func AppAbilities() []string {
	return []string{
		"catalog:read", "ratings:write", "status:write", "lists:write", "profile:write", "suggestions:write",
	}
}

// CleanDeviceName strips control runes and trims; ok is false when nothing is left or it is too long.
func CleanDeviceName(raw string) (string, bool) {
	name := strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, raw))
	n := len([]rune(name))

	return name, n >= 1 && n <= MaxDeviceName
}

func live(t auth.Token, now time.Time) bool {
	return t.RevokedAt == nil && (t.ExpiresAt == nil || t.ExpiresAt.After(now))
}

// IssueAppToken creates a 60-day app token for u named after the device. When u already holds
// MaxAppTokens live tokens, the least recently used are revoked first (SDD 03 §7.4).
func (a *Accounts) IssueAppToken(ctx context.Context, u models.User, device string) (string, auth.Token, error) {
	uid := strconv.FormatInt(u.ID, 10)

	all, err := a.Tokens.List(ctx, uid)
	if err != nil {
		return "", auth.Token{}, err
	}

	var liveTokens []auth.Token

	// List is newest first; reversing gives creation order, which breaks ties of never-used tokens.
	for _, t := range slices.Backward(all) {
		if live(t, a.now()) {
			liveTokens = append(liveTokens, t)
		}
	}

	lastUsed := func(t auth.Token) time.Time {
		if t.LastUsedAt == nil {
			return time.Time{}
		}

		return *t.LastUsedAt
	}

	slices.SortStableFunc(liveTokens, func(x, y auth.Token) int { return lastUsed(x).Compare(lastUsed(y)) })

	for len(liveTokens) >= MaxAppTokens {
		if _, err := a.Tokens.RevokeFor(ctx, uid, liveTokens[0].ID); err != nil {
			return "", auth.Token{}, err
		}

		liveTokens = liveTokens[1:]
	}

	return a.Tokens.Create(ctx, uid, device, AppAbilities(), AppTokenTTL)
}

// ErrTokenAlreadyRotated is returned when the token presented for rotation
// was revoked by a concurrent rotation or revoke.
var ErrTokenAlreadyRotated = errors.New("accounts: token already rotated")

// RotateAppToken revokes current and issues a replacement named like it.
// The conditional revoke runs first so only one of two concurrent rotations
// with the same token wins: a stolen token cannot mint a second live token
// next to the device's (security review, P1.C.4.1). A failure after the
// revoke leaves the device without a token; it signs in again.
func (a *Accounts) RotateAppToken(ctx context.Context, u models.User, current auth.Token) (string, auth.Token, error) {
	uid := strconv.FormatInt(u.ID, 10)

	ok, err := a.Tokens.RevokeFor(ctx, uid, current.ID)
	if err != nil {
		return "", auth.Token{}, err
	}

	if !ok {
		return "", auth.Token{}, ErrTokenAlreadyRotated
	}

	return a.Tokens.Create(ctx, uid, current.Name, AppAbilities(), AppTokenTTL)
}

// LiveAppTokens lists u's unrevoked, unexpired tokens, newest first.
func (a *Accounts) LiveAppTokens(ctx context.Context, u models.User) ([]auth.Token, error) {
	all, err := a.Tokens.List(ctx, strconv.FormatInt(u.ID, 10))
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(all, func(t auth.Token) bool { return !live(t, a.now()) }), nil
}
