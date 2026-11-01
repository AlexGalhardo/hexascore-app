package accounts

import (
	"context"

	"github.com/AlexGalhardo/vukem/auth"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// UserProvider is the ORM provider with one extra rule: a seed user (users.is_seed) is reported as not found, so
// password login, remember-me, app and API tokens, session reloads and password reset all fail closed with the same
// answer as an unknown account. Fake seed users must never sign in (owner decision 2026-11-01).
type UserProvider struct {
	*auth.ORMProvider[models.User]
}

// ByID implements auth.UserProvider.
func (p UserProvider) ByID(ctx context.Context, id string) (models.User, error) {
	return notSeed(p.ORMProvider.ByID(ctx, id))
}

// ByCredentials implements auth.UserProvider.
func (p UserProvider) ByCredentials(ctx context.Context, creds map[string]string) (models.User, error) {
	return notSeed(p.ORMProvider.ByCredentials(ctx, creds))
}

func notSeed(u models.User, err error) (models.User, error) {
	if err == nil && u.IsSeed {
		return models.User{}, auth.ErrUserNotFound
	}

	return u, err
}
