package billing

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/contrib/stripe"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/config"
)

// Canceller ends a user's Stripe subscriptions when the account is deleted
// (accounts.SubscriptionCanceller). It needs the Stripe key only.
type Canceller struct {
	DB     *database.DB
	Stripe *stripe.Client
}

// NewCanceller builds a Canceller; Stripe stays nil while STRIPE_KEY is unset.
func NewCanceller(app *vukem.App, db *database.DB) (*Canceller, error) {
	c, err := newClient(app, config.LoadBilling(app.Source))
	if err != nil {
		return nil, err
	}

	return &Canceller{DB: db, Stripe: c}, nil
}

// CancelLive cancels every Stripe subscription of the user that can still
// bill (active or past_due). A Stripe failure is returned so the deletion
// stays scheduled and retries; a subscription Stripe no longer knows counts as
// cancelled.
func (c *Canceller) CancelLive(ctx context.Context, userID int64) error {
	subs, err := orm.QueryOf[models.Subscription](c.DB).Where("user_id", userID).
		Where("provider", "stripe").WhereIn("state", []any{StateActive, StatePastDue}).Get(ctx)
	if err != nil || len(subs) == 0 {
		return err
	}

	if c.Stripe == nil {
		return ErrNotConfigured
	}

	for _, s := range subs {
		if _, err := c.Stripe.CancelSubscription(ctx, s.StripeSubscriptionID); err != nil {
			if se, ok := errors.AsType[*stripe.Error](err); !ok || se.Status != http.StatusNotFound {
				return fmt.Errorf("cancel stripe subscription %s: %w", s.StripeSubscriptionID, err)
			}
		}

		_, err := orm.QueryOf[models.Subscription](c.DB).Where("id", s.ID).
			Update(ctx, orm.Set{"state": StateCanceled, "status": "canceled"})
		if err != nil {
			return err
		}
	}

	return nil
}
