package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/AlexGalhardo/vukem/contrib/stripe"
)

// workaround(vukem): VF-NEW-P1.H.3.1 — contrib/stripe has no subscription list, no checkout session
// list/expire and no Subscription.Created, so these go through Client.Call.

// getSubscription is GetSubscription plus the subscription's created time (unix seconds).
func getSubscription(ctx context.Context, c *stripe.Client, id string) (stripe.Subscription, int64, error) {
	var raw json.RawMessage
	if err := c.Call(ctx, http.MethodGet, "/v1/subscriptions/"+url.PathEscape(id), nil, &raw); err != nil {
		return stripe.Subscription{}, 0, err
	}

	var sub stripe.Subscription
	if err := json.Unmarshal(raw, &sub); err != nil {
		return sub, 0, fmt.Errorf("decode subscription %s: %w", id, err)
	}

	// Decoded separately: Subscription's UnmarshalJSON would swallow an embedding struct's extra field.
	var meta struct {
		Created int64 `json:"created"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return sub, 0, fmt.Errorf("decode subscription %s: %w", id, err)
	}

	return sub, meta.Created, nil
}

// customerSubscriptions lists every subscription of a Stripe customer, whatever its status.
func customerSubscriptions(ctx context.Context, c *stripe.Client, customer string) ([]stripe.Subscription, error) {
	var out []stripe.Subscription

	p := stripe.Params{"customer": customer, "status": "all", "limit": 100}
	for {
		var page stripe.List[stripe.Subscription]
		if err := c.Call(ctx, http.MethodGet, "/v1/subscriptions", p, &page); err != nil {
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
	var page stripe.List[stripe.CheckoutSession]
	if err := c.Call(ctx, http.MethodGet, "/v1/checkout/sessions",
		stripe.Params{"customer": customer, "status": "open", "limit": 100}, &page); err != nil {
		return fmt.Errorf("list open checkout sessions: %w", err)
	}

	for _, cs := range page.Data {
		if cs.Mode != "subscription" {
			continue
		}

		err := c.Call(ctx, http.MethodPost, "/v1/checkout/sessions/"+url.PathEscape(cs.ID)+"/expire", nil, nil)
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
