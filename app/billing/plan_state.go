package billing

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Subscription states (SDD 05 section 7.5).
const (
	StateActive   = "active"
	StatePastDue  = "past_due"
	StateCanceled = "canceled"
	StateExpired  = "expired"
)

// StateOf maps a Stripe subscription status to a plan state. Unknown statuses
// grant nothing.
func StateOf(status string) string {
	switch status {
	case "active", "trialing":
		return StateActive
	case "canceled", "incomplete_expired":
		return StateCanceled
	default:
		return StatePastDue
	}
}

// EffectivePlan is the API plan a subscription row grants at now: its own plan
// while the state is active and the paid period has not ended, else free.
func EffectivePlan(s models.Subscription, now time.Time) string {
	if s.State != StateActive {
		return api.PlanFree
	}

	if s.CurrentPeriodEnd.Valid && !s.CurrentPeriodEnd.V.After(now) {
		return api.PlanFree
	}

	if _, ok := api.PlanByID(s.Plan); !ok {
		return api.PlanFree
	}

	return s.Plan
}

// PlanResolver is the api.PlanResolver backed by the subscriptions table: one
// indexed lookup of the user's live row per request.
func PlanResolver(db *database.DB, now func() time.Time) api.PlanResolver {
	return func(ctx context.Context, userID int64) (api.Plan, error) {
		sub, err := orm.QueryOf[models.Subscription](db).
			Where("user_id", userID).Where("state", StateActive).First(ctx)
		id := api.PlanFree

		switch {
		case errors.Is(err, orm.ErrNotFound), errors.Is(err, sql.ErrNoRows):
		case err != nil:
			return api.Plan{}, err
		default:
			id = EffectivePlan(sub, now())
		}

		p, _ := api.PlanByID(id)

		return p, nil
	}
}
