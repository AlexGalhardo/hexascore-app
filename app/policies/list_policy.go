// Package policies holds the authorization policies of the app gate (vukem authz, SDD 13).
package policies

import (
	"context"

	"github.com/AlexGalhardo/vukem/authz"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// ListPolicy is SDD 14 §6.1: everyone else sees a list only when it is public, not hidden and its owner is not
// private; every other denial answers 404 so a private list is indistinguishable from a missing one.
type ListPolicy struct{}

// Create allows any signed-in user (u is nil for anonymous callers).
func (ListPolicy) Create(_ context.Context, u *models.User) bool { return u != nil }

// View allows the owner, or anyone when the list is public and visible.
func (ListPolicy) View(_ context.Context, u *models.User, l *models.List) authz.Result {
	if owns(u, l) || (l.IsPublic && !l.Hidden && !l.OwnerPrivate) {
		return authz.Allow()
	}

	return authz.DenyAsNotFound()
}

// Update allows the owner only.
func (ListPolicy) Update(_ context.Context, u *models.User, l *models.List) authz.Result {
	return ownerOnly(u, l)
}

// Delete allows the owner only.
func (ListPolicy) Delete(_ context.Context, u *models.User, l *models.List) authz.Result {
	return ownerOnly(u, l)
}

// AddItem allows the owner only.
func (ListPolicy) AddItem(_ context.Context, u *models.User, l *models.List) authz.Result {
	return ownerOnly(u, l)
}

// Reorder allows the owner only.
func (ListPolicy) Reorder(_ context.Context, u *models.User, l *models.List) authz.Result {
	return ownerOnly(u, l)
}

func owns(u *models.User, l *models.List) bool { return u != nil && u.ID == l.UserID }

func ownerOnly(u *models.User, l *models.List) authz.Result {
	if owns(u, l) {
		return authz.Allow()
	}

	return authz.DenyAsNotFound()
}
