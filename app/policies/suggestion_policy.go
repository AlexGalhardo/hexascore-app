package policies

import (
	"context"

	"github.com/AlexGalhardo/vukem/authz"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// SuggestionPolicy is SDD 04 FR-04-12: the author and admins see a suggestion; everyone else, anonymous included,
// gets 404 so another user's suggestion is indistinguishable from a missing one. Only the author acts on it.
type SuggestionPolicy struct{}

// View allows the author and admins.
func (SuggestionPolicy) View(_ context.Context, u *models.User, s *models.Suggestion) authz.Result {
	if u != nil && (u.Role == models.RoleAdmin || authored(u, s)) {
		return authz.Allow()
	}

	return authz.DenyAsNotFound()
}

// Submit allows the author only.
func (SuggestionPolicy) Submit(_ context.Context, u *models.User, s *models.Suggestion) authz.Result {
	return authorOnly(u, s)
}

// Withdraw allows the author only.
func (SuggestionPolicy) Withdraw(_ context.Context, u *models.User, s *models.Suggestion) authz.Result {
	return authorOnly(u, s)
}

func authored(u *models.User, s *models.Suggestion) bool {
	return s.AuthorID.Valid && s.AuthorID.Int64 == u.ID
}

func authorOnly(u *models.User, s *models.Suggestion) authz.Result {
	if u != nil && authored(u, s) {
		return authz.Allow()
	}

	return authz.DenyAsNotFound()
}
