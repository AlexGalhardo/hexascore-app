package billing

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/AlexGalhardo/vukem/contrib/stripe"
)

// customerSubscriptions lists every subscription of a Stripe customer, whatever its status.
func customerSubscriptions(ctx context.Context, c *stripe.Client, customer string) ([]stripe.Subscription, error) {
	var out []stripe.Subscription

	p := stripe.Params{"customer": customer, "status": "all", "limit": 100}
	for {
		page, err := c.ListSubscriptions(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("list stripe subscriptions: %w", err)
		}

		out = append(out, page.Data...)
		if !page.HasMore || len(page.Data) == 0 {
			return out, nil
		}

		p["starting_after"] = page.Data[len(page.Data)-1].ID
	}
}

// expireOpenCheckouts expires the customer's open subscription Checkout Sessions, so only the session about to be
// created can still start a subscription (two tabs would otherwise create two).
func expireOpenCheckouts(ctx context.Context, c *stripe.Client, customer string) error {
	page, err := c.ListCheckoutSessions(ctx, stripe.Params{"customer": customer, "status": "open", "limit": 100})
	if err != nil {
		return fmt.Errorf("list open checkout sessions: %w", err)
	}

	for _, cs := range page.Data {
		if cs.Mode != "subscription" {
			continue
		}

		_, err := c.ExpireCheckoutSession(ctx, cs.ID)
		if err != nil && !isNotFoundOrNotOpen(err) {
			return fmt.Errorf("expire checkout session %s: %w", cs.ID, err)
		}
	}

	return nil
}

// cancelSubscription cancels at Stripe; one Stripe no longer knows counts as cancelled.
func cancelSubscription(ctx context.Context, c *stripe.Client, id string) error {
	if _, err := c.CancelSubscription(ctx, id); err != nil {
		if se, ok := errors.AsType[*stripe.Error](err); !ok || se.Status != http.StatusNotFound {
			return fmt.Errorf("cancel stripe subscription %s: %w", id, err)
		}
	}

	return nil
}

// isNotFoundOrNotOpen is a session already gone or completed meanwhile (404 or 400 from /expire).
func isNotFoundOrNotOpen(err error) bool {
	se, ok := errors.AsType[*stripe.Error](err)

	return ok && (se.Status == http.StatusNotFound || se.Status == http.StatusBadRequest)
}
