package admin

import (
	"context"
	"time"

	"github.com/AlexGalhardo/vukem/authz"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Moderation abilities (SDD 04 §6).
const (
	AbilityModerate         = "moderate"
	AbilityApplySuggestion  = "applySuggestion"
	AbilityViewCatalogAudit = "viewCatalogAudit"
	AbilityRevertCatalog    = "revertCatalog"
	AbilityLockCatalog      = "lockCatalog"
	AbilityRunIngestion     = "runIngestion"
)

// DefineModeration registers the moderation abilities. Unlike the panel gate they answer 403, not 404: the pages
// only exist behind the flag-gated role, and a refused member should know it. Applying or rejecting a suggestion,
// revert, locks and the audit log stay admin-only (owner Q-A15: moderators only review); moderate, which covers the
// queue, claims and recommendations, also goes to the moderator role while trustedModerators is on (D-15). Banned or unverified
// accounts, a missing user and unknown roles are always refused.
func DefineModeration(
	g *authz.Gate[*models.User],
	now func() time.Time,
	trustedModerators func(context.Context) bool,
) {
	adminOnly := func(_ context.Context, u *models.User) authz.Result {
		if IsAdmin(u, now()) {
			return authz.Allow()
		}

		return authz.Deny()
	}
	g.DefineResult(AbilityApplySuggestion, adminOnly)
	g.DefineResult(AbilityViewCatalogAudit, adminOnly)
	g.DefineResult(AbilityRevertCatalog, adminOnly)
	g.DefineResult(AbilityLockCatalog, adminOnly)
	g.DefineResult(AbilityRunIngestion, adminOnly)
	g.DefineResult(AbilityModerate, func(ctx context.Context, u *models.User) authz.Result {
		if IsAdmin(u, now()) {
			return authz.Allow()
		}
		if u != nil && u.Role == models.RoleModerator && !u.IsBanned(now()) && u.EmailVerified() &&
			trustedModerators(ctx) {
			return authz.Allow()
		}

		return authz.Deny()
	})
}
