package admin

import (
	"time"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/authz"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Resources lists the panel's resources; vukem make:admin-resource appends
// here. The scoring tables exist only on PostgreSQL (VF-033), so the SQLite
// scaffold tests get the users resource alone.
func Resources(acc *accounts.Accounts, postgres bool) []vadmin.AnyResource {
	out := []vadmin.AnyResource{UsersResource(acc)}
	if postgres {
		out = append(out, ScoreBurstsResource(), RatingFlagsResource(), ScoreActionsResource(),
			ScoreFormulaVersionsResource(),
			ReportsResource(), ContactMessagesResource(), ItemsResource())
	}

	return out
}

// Policies lists what admins may do per model; vukem make:admin-resource
// appends here. A resource without a policy denies every action.
func Policies(now func() time.Time) []authz.Policy {
	return []authz.Policy{
		authz.For[*models.User](UserPolicy{now: now}),
		authz.For[*models.ScoreBurst](readOnlyPolicy[models.ScoreBurst]{now: now, detail: true}),
		authz.For[*models.RatingFlag](readOnlyPolicy[models.RatingFlag]{now: now}),
		authz.For[*models.ScoreAction](readOnlyPolicy[models.ScoreAction]{now: now, detail: true}),
		authz.For[*models.ScoreFormulaVersion](readOnlyPolicy[models.ScoreFormulaVersion]{now: now, detail: true}),
		authz.For[*models.Report](readOnlyPolicy[models.Report]{now: now, detail: true}),
		authz.For[*models.Item](readOnlyPolicy[models.Item]{now: now, detail: true}),
		authz.For[*models.ContactMessage](contactPolicy{now: now}),
	}
}
