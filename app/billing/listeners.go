package billing

import (
	"context"
	"strconv"

	"github.com/AlexGalhardo/vukem/contrib/stripe"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// invoice is the part of a Stripe invoice the sync needs.
type invoice struct {
	// Subscription is set by API versions before 2025-03-31.
	Subscription string `json:"subscription"`
	Parent       struct {
		SubscriptionDetails struct {
			Subscription string `json:"subscription"`
		} `json:"subscription_details"`
	} `json:"parent"`
}

// HandleWebhook syncs subscriptions from Stripe events (register it with
// events.Listen). An error answers 500, so Stripe delivers the event again.
func (b *Billing) HandleWebhook(ctx context.Context, r stripe.Received) error {
	ev := r.Event

	switch ev.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded":
		cs, err := stripe.Decode[stripe.CheckoutSession](ev)
		if err != nil {
			return err
		}

		if cs.Mode != "subscription" {
			return nil
		}

		if err := b.linkCustomer(ctx, cs); err != nil {
			return err
		}

		// Delayed payment methods complete as "unpaid": access starts with
		// checkout.session.async_payment_succeeded.
		if cs.PaymentStatus != "paid" && cs.PaymentStatus != "no_payment_required" {
			return nil
		}

		return b.SyncSubscription(ctx, cs.Subscription)

	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted",
		"customer.subscription.paused", "customer.subscription.resumed":
		sub, err := stripe.Decode[stripe.Subscription](ev)
		if err != nil {
			return err
		}

		return b.SyncSubscription(ctx, sub.ID)

	case "invoice.paid", "invoice.payment_failed":
		inv, err := stripe.Decode[invoice](ev)
		if err != nil {
			return err
		}

		id := inv.Parent.SubscriptionDetails.Subscription
		if id == "" {
			id = inv.Subscription
		}

		return b.SyncSubscription(ctx, id)
	}

	return nil
}

// linkCustomer records the Checkout customer on the user named by
// client_reference_id when that user has none yet: a session can never
// move another user's customer.
func (b *Billing) linkCustomer(ctx context.Context, cs stripe.CheckoutSession) error {
	id, err := strconv.ParseInt(cs.ClientRefID, 10, 64)
	if err != nil || cs.Customer == "" {
		return nil //nolint:nilerr // sessions created elsewhere have no user id
	}

	taken, err := orm.QueryOf[models.Billable](b.DB).Where("stripe_customer_id", cs.Customer).Exists(ctx)
	if err != nil || taken {
		return err
	}

	_, err = orm.QueryOf[models.Billable](b.DB).Where("id", id).WhereNull("stripe_customer_id").
		Update(ctx, orm.Set{"stripe_customer_id": cs.Customer})

	return err
}
