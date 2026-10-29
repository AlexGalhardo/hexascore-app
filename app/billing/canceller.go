package billing

import (
	"context"
	"errors"

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

// CancelLive cancels every subscription of the user's Stripe customer that
// Stripe may still bill, read from Stripe itself (status=all) rather than
// from the local rows, which can lag or be retired. Any Stripe failure,
// including the list call, is returned so the deletion stays scheduled and
// retries.
func (c *Canceller) CancelLive(ctx context.Context, userID int64) error {
	u, err := orm.QueryOf[models.Billable](c.DB).Where("id", userID).First(ctx)
	if errors.Is(err, orm.ErrNotFound) || (err == nil && !u.StripeCustomerID.Valid) {
		return nil // never reached Stripe: nothing can bill
	}

	if err != nil {
		return err
	}

	if c.Stripe == nil {
		return ErrNotConfigured
	}

	subs, err := customerSubscriptions(ctx, c.Stripe, u.StripeCustomerID.V)
	if err != nil {
		return err
	}

	for _, s := range subs {
		if s.Status == "canceled" || s.Status == "incomplete_expired" {
			continue
		}

		if err := cancelSubscription(ctx, c.Stripe, s.ID); err != nil {
			return err
		}
	}

	_, err = orm.QueryOf[models.Subscription](c.DB).Where("user_id", userID).Where("provider", "stripe").
		Update(ctx, orm.Set{"state": StateCanceled, "status": "canceled"})

	return err
}
