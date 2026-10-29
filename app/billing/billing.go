// Package billing sells subscriptions with Stripe Checkout, lets
// customers manage them in the Stripe Billing Portal and keeps the
// subscriptions table in sync from Stripe webhooks (vukem
// install:billing). It is your code: change the rules freely.
//
// Access is granted from webhooks only, never from the Checkout success
// page: develop against a Stripe sandbox (`stripe sandbox create`) and
// forward its events with
//
//	stripe listen --forward-to localhost:3000/stripe/webhook
package billing

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/contrib/stripe"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/config"
)

// WebhookPath is where Stripe sends events; register it in the Dashboard
// (Developers → Webhooks) or with stripe listen.
const WebhookPath = "/stripe/webhook"

var (
	// ErrNotConfigured answers billing requests while STRIPE_KEY is unset
	// (development only: production does not boot without it).
	ErrNotConfigured = errs.New(http.StatusServiceUnavailable, "Billing is not configured.")
	// ErrSubscriptionRequired answers JSON clients without a subscription.
	ErrSubscriptionRequired = errs.New(http.StatusPaymentRequired, "A subscription is required.")

	errAlreadySubscribed = errs.New(
		http.StatusConflict,
		"You already have a subscription: manage it in the billing portal.",
	)
	errUnsupportedCurrency = errs.New(http.StatusUnprocessableEntity, "Unsupported currency.")

	errNoCustomer = errs.New(http.StatusConflict, "You have no billing account yet: choose a plan first.")
)

// Billing holds the Stripe client and the billing rules.
type Billing struct {
	DB *database.DB
	// Stripe is nil while STRIPE_KEY is unset (development only).
	Stripe *stripe.Client
	Config config.Billing
	// URL is the public base URL Stripe sends customers back to (APP_URL).
	URL string
	Log *slog.Logger

	webhook http.Handler
}

// New builds Billing from the app. Missing settings stop production at
// boot; elsewhere billing answers 503 until they are set.
func New(app *vukem.App) (*Billing, error) {
	db, err := app.DB()
	if err != nil {
		return nil, err
	}

	cfg := config.LoadBilling(app.Source)
	if err := cfg.Validate(); err != nil {
		if app.Config.IsProduction() {
			return nil, fmt.Errorf("billing: %w", err)
		}

		if app.Kind == vukem.KindWeb {
			app.Log.Warn("billing is not configured", "err", err)
		}
	}

	b := &Billing{
		DB: db, Config: cfg, Log: app.Log,
		URL: strings.TrimSuffix(app.Config.URL, "/"),
	}

	if b.Stripe, err = newClient(app, cfg); err != nil {
		return nil, err
	}

	if !cfg.WebhookSecret.IsZero() {
		c, err := app.Cache()
		if err != nil {
			return nil, err
		}

		// The cache remembers processed event ids: use a shared store
		// (CACHE_STORE=database or Redis) when several instances run.
		b.webhook, err = stripe.WebhookHandler(stripe.WebhookConfig{
			Secret: cfg.WebhookSecret.Reveal(), Store: c.Store(), Bus: app.Events, Log: app.Log,
		})
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

// newClient builds the Stripe client, nil while STRIPE_KEY is unset.
func newClient(app *vukem.App, cfg config.Billing) (*stripe.Client, error) {
	if cfg.Key.IsZero() {
		return nil, nil //nolint:nilnil // no key (development): callers answer ErrNotConfigured
	}

	var opts []stripe.Option
	// STRIPE_API_BASE points the client at the fake API of the tests;
	// it is ignored outside them, so the key only ever goes to Stripe.
	if base := app.Source.Get("STRIPE_API_BASE"); base != "" && app.Kind == vukem.KindTest {
		opts = append(opts, stripe.WithBaseURL(base))
	}

	c, err := stripe.New(cfg.Key.Reveal(), opts...)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(cfg.Key.Reveal(), "sk_") {
		app.Log.Warn("STRIPE_KEY is a secret key: use a restricted key (rk_) with only the billing permissions")
	}

	return c, nil
}

// Webhook serves Stripe webhooks at WebhookPath ahead of the router
// stack: Stripe is a server, not a browser, so the endpoint has no session
// or CSRF check. A valid Stripe-Signature is its authentication, and each
// event id is processed once. Register it with r.UseServer.
func (b *Billing) Webhook() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != WebhookPath {
				next.ServeHTTP(w, r)
				return
			}

			if b.webhook == nil { // no STRIPE_WEBHOOK_SECRET: refuse (fail closed)
				http.Error(w, "webhooks are not configured", http.StatusServiceUnavailable)
				return
			}

			b.webhook.ServeHTTP(w, r)
		})
	}
}

// Current returns the user's subscription that grants access, else the
// latest one; ok is false when the user never subscribed.
func (b *Billing) Current(ctx context.Context, userID int64) (sub models.Subscription, ok bool, err error) {
	subs, err := orm.QueryOf[models.Subscription](b.DB).Where("user_id", userID).Latest().Get(ctx)
	if err != nil || len(subs) == 0 {
		return models.Subscription{}, false, err
	}

	for i := range subs {
		if subs[i].Valid() {
			return subs[i], true, nil
		}
	}

	return subs[0], true, nil
}

// Customer returns the Stripe customer of user, creating it on first use.
func (b *Billing) Customer(ctx context.Context, user models.User) (string, error) {
	if b.Stripe == nil {
		return "", ErrNotConfigured
	}

	bill, err := orm.QueryOf[models.Billable](b.DB).Find(ctx, user.ID)
	if err != nil {
		return "", err
	}

	if bill.StripeCustomerID.Valid {
		return bill.StripeCustomerID.V, nil
	}

	cus, err := b.Stripe.CreateCustomer(ctx, stripe.Params{
		"email": user.Email, "name": user.Name,
		"metadata": map[string]string{"user_id": user.AuthID()},
	})
	if err != nil {
		return "", err
	}

	n, err := orm.QueryOf[models.Billable](b.DB).Where("id", user.ID).WhereNull("stripe_customer_id").
		Update(ctx, orm.Set{"stripe_customer_id": cus.ID})
	if err != nil || n > 0 {
		return cus.ID, err
	}

	// A concurrent request linked a customer first: use that one.
	bill, err = orm.QueryOf[models.Billable](b.DB).Find(ctx, user.ID)

	return bill.StripeCustomerID.V, err
}

// Checkout starts a Stripe Checkout Session in subscription mode for the
// plan named plan and returns its URL. Access starts when the webhook
// arrives, not when the customer comes back.
func (b *Billing) Checkout(ctx context.Context, user models.User, plan, currency string) (string, error) {
	currency, ok := config.Currency(cmp.Or(strings.TrimSpace(currency), config.Currencies[0]))
	if !ok {
		return "", errUnsupportedCurrency
	}

	p, ok := b.Config.Plan(plan)
	if !ok || p.Price == "" {
		return "", errs.ErrNotFound
	}

	if b.Stripe == nil {
		return "", ErrNotConfigured
	}

	if sub, ok, err := b.Current(ctx, user.ID); err != nil {
		return "", err
	} else if ok && sub.Valid() {
		return "", errAlreadySubscribed // plan changes happen in the portal
	}

	customer, err := b.Customer(ctx, user)
	if err != nil {
		return "", err
	}

	if err := expireOpenCheckouts(ctx, b.Stripe, customer); err != nil {
		return "", err
	}

	// No payment_method_types: Stripe picks the payment methods enabled
	// in the Dashboard.
	cs, err := b.Stripe.CreateCheckoutSession(ctx, stripe.Params{
		"mode":                "subscription",
		"customer":            customer,
		"currency":            currency,
		"client_reference_id": user.AuthID(),
		"line_items":          []stripe.Params{{"price": p.Price, "quantity": 1}},
		"subscription_data":   stripe.Params{"metadata": map[string]string{"user_id": user.AuthID()}},
		"success_url":         b.URL + "/billing?checkout=success",
		"cancel_url":          b.URL + "/billing?checkout=canceled",
	})
	if err != nil {
		return "", err
	}

	return cs.URL, nil
}

// Portal returns a Stripe Billing Portal URL where the user changes plan,
// updates payment details, sees invoices or cancels.
func (b *Billing) Portal(ctx context.Context, user models.User) (string, error) {
	if b.Stripe == nil {
		return "", ErrNotConfigured
	}

	bill, err := orm.QueryOf[models.Billable](b.DB).Find(ctx, user.ID)
	if err != nil {
		return "", err
	}

	if !bill.StripeCustomerID.Valid {
		return "", errNoCustomer
	}

	ps, err := b.Stripe.CreatePortalSession(ctx, bill.StripeCustomerID.V, b.URL+"/billing")
	if err != nil {
		return "", err
	}

	return ps.URL, nil
}

// SyncSubscription stores the current state of subscription id. It is
// fetched from Stripe rather than read from the event, because events can
// arrive out of order and the latest state is what counts.
func (b *Billing) SyncSubscription(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}

	if b.Stripe == nil {
		return ErrNotConfigured
	}

	sub, created, err := getSubscription(ctx, b.Stripe, id)
	if err != nil {
		return err
	}

	user, err := orm.QueryOf[models.Billable](b.DB).Where("stripe_customer_id", sub.Customer).First(ctx)
	if errors.Is(err, orm.ErrNotFound) {
		b.Log.WarnContext(ctx, "stripe subscription of an unknown customer", "subscription", sub.ID)
		return nil
	}

	if err != nil {
		return err
	}

	row := models.Subscription{
		UserID: user.ID, StripeSubscriptionID: sub.ID, Status: sub.Status,
		Provider: "stripe", State: StateOf(sub.Status), CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
	}

	if len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		row.PriceID = item.Price.ID

		if p, found := b.Config.PlanForPrice(row.PriceID); found {
			row.Plan = p.Name
		}

		if item.CurrentPeriodEnd > 0 {
			row.CurrentPeriodEnd = orm.NullTime{Null: sql.Null[time.Time]{
				V: time.Unix(item.CurrentPeriodEnd, 0).UTC(), Valid: true,
			}}
		}
	}

	if row.State == StateActive {
		if row.State, err = b.supersede(ctx, row.UserID, sub.ID, created); err != nil {
			return err
		}

		if row.State == StateCanceled {
			row.Status = "canceled"
		}
	}

	return b.DB.Tx(ctx, func(tx *database.Tx) error {
		var periodEnd any
		if row.CurrentPeriodEnd.Valid {
			periodEnd = row.CurrentPeriodEnd.V
		}

		// One live row per user: supersede already cancelled the other live
		// subscriptions at Stripe; their rows follow.
		if row.State == StateActive {
			_, err := orm.QueryOf[models.Subscription](tx).Where("user_id", row.UserID).
				Where("state", StateActive).Where("stripe_subscription_id", "!=", sub.ID).
				Update(ctx, orm.Set{"state": StateCanceled})
			if err != nil {
				return err
			}
		}

		n, err := orm.QueryOf[models.Subscription](tx).Where("stripe_subscription_id", sub.ID).Update(ctx, orm.Set{
			"user_id": row.UserID, "status": row.Status, "price_id": row.PriceID,
			"plan": row.Plan, "provider": row.Provider, "state": row.State,
			"current_period_end": periodEnd, "cancel_at_period_end": row.CancelAtPeriodEnd,
		})
		if err != nil || n > 0 {
			return err
		}

		return orm.Create(ctx, tx, &row)
	})
}

// Subscribed lets through users with an active or trialing subscription;
// place it after auth.Required. Others are sent to /billing (402 for JSON
// clients). A database error blocks too (fail closed).
func (b *Billing) Subscribed() func(http.Handler) http.Handler {
	return b.require(func(models.Subscription) bool { return true })
}

// OnPlan is Subscribed limited to the plans named plans ("pro").
func (b *Billing) OnPlan(plans ...string) func(http.Handler) http.Handler {
	prices := map[string]bool{}

	for _, name := range plans {
		if p, ok := b.Config.Plan(name); ok && p.Price != "" {
			prices[p.Price] = true
		}
	}

	return b.require(func(s models.Subscription) bool { return prices[s.PriceID] })
}

func (b *Billing) require(allowed func(models.Subscription) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if b.allows(r.Context(), allowed) {
				next.ServeHTTP(w, r)
				return
			}

			c := web.NewCtx(w, r, web.ServicesFrom(r.Context()))
			if c.WantsJSON() {
				c.HandleError(ErrSubscriptionRequired)
				return
			}

			_ = c.Flash("status", "Choose a plan to continue.")
			_ = c.Redirect("/billing")
		})
	}
}

func (b *Billing) allows(ctx context.Context, allowed func(models.Subscription) bool) bool {
	u, err := auth.UserFrom[models.User](ctx, "web")
	if err != nil {
		return false
	}

	sub, ok, err := b.Current(ctx, u.ID)
	if err != nil {
		b.Log.ErrorContext(ctx, "billing: load subscription", "user_id", u.ID, "err", err)
		return false
	}

	return ok && sub.Valid() && allowed(sub)
}

// supersede keeps one billing subscription per user when two go live (for
// example two Checkout tabs): the newest by Stripe created time wins and the
// other is cancelled at Stripe, so the user never pays for a subscription the
// app ignores. It returns the state the synced subscription keeps.
func (b *Billing) supersede(ctx context.Context, userID int64, id string, created int64) (string, error) {
	others, err := orm.QueryOf[models.Subscription](b.DB).Where("user_id", userID).
		Where("state", StateActive).Where("stripe_subscription_id", "!=", id).Get(ctx)
	if err != nil {
		return "", err
	}

	for _, o := range others {
		other, otherCreated, err := getSubscription(ctx, b.Stripe, o.StripeSubscriptionID)
		if err != nil {
			return "", err
		}

		if StateOf(other.Status) != StateActive {
			continue // stale local row: the transaction retires it
		}

		if otherCreated > created || (otherCreated == created && other.ID > id) {
			b.Log.WarnContext(ctx, "cancelling a superseded stripe subscription", "subscription", id, "kept", other.ID)

			return StateCanceled, cancelSubscription(ctx, b.Stripe, id)
		}

		b.Log.WarnContext(ctx, "cancelling a superseded stripe subscription", "subscription", other.ID, "kept", id)

		if err := cancelSubscription(ctx, b.Stripe, other.ID); err != nil {
			return "", err
		}
	}

	return StateActive, nil
}
