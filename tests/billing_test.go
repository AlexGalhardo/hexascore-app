package tests

import (
	"bytes"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/contrib/stripe"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/vukemtest"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/billing"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

const (
	webhookSecret = "whsec_test_vukem"
	stripeKey     = "rk_test_vukem"
	// The fixtures every test shares: one customer, one subscription.
	testCustomer     = "cus_1"
	testSubscription = "sub_1"
)

// fakeStripe stands in for api.stripe.com: it serves the objects the app
// asks for and records what it was sent. Tests never reach the network.
type fakeStripe struct {
	cancels    []string
	failCancel bool
	failList   bool
	expired    []string
	// openSessions are the open Checkout Sessions GET /v1/checkout/sessions serves.
	openSessions []map[string]any
	t            *testing.T
	srv          *httptest.Server

	mu        sync.Mutex
	customers int
	checkouts []url.Values
	portals   []url.Values
	subs      map[string]map[string]any
	subReads  int
}

func newFakeStripe(t *testing.T) *fakeStripe {
	t.Helper()

	f := &fakeStripe{t: t, subs: map[string]map[string]any{}}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/customers", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.customers++
		id := "cus_" + strconv.Itoa(f.customers)
		f.mu.Unlock()

		f.json(w, map[string]any{"id": id, "email": r.PostFormValue("email")})
	})
	mux.HandleFunc("POST /v1/checkout/sessions", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()

		f.mu.Lock()
		f.checkouts = append(f.checkouts, r.PostForm)
		f.mu.Unlock()

		f.json(w, map[string]any{"id": "cs_test_1", "url": "https://checkout.stripe.com/c/pay/cs_test_1"})
	})
	mux.HandleFunc("POST /v1/billing_portal/sessions", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()

		f.mu.Lock()
		f.portals = append(f.portals, r.PostForm)
		f.mu.Unlock()

		f.json(w, map[string]any{"id": "bps_1", "url": "https://billing.stripe.com/p/session/test_1"})
	})
	mux.HandleFunc("DELETE /v1/subscriptions/{id}", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.cancels = append(f.cancels, r.PathValue("id"))
		fail := f.failCancel
		if sub, ok := f.subs[r.PathValue("id")]; ok && !fail {
			sub["status"] = "canceled"
		}
		f.mu.Unlock()

		if fail {
			w.WriteHeader(http.StatusInternalServerError)
			f.json(w, map[string]any{"error": map[string]any{"type": "api_error"}})

			return
		}

		f.json(w, map[string]any{"id": r.PathValue("id"), "status": "canceled"})
	})
	mux.HandleFunc("GET /v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()

		if f.failList {
			w.WriteHeader(http.StatusInternalServerError)
			f.json(w, map[string]any{"error": map[string]any{"type": "api_error"}})

			return
		}

		data := []any{}
		for _, id := range slices.Sorted(maps.Keys(f.subs)) {
			if f.subs[id]["customer"] == r.URL.Query().Get("customer") {
				data = append(data, f.subs[id])
			}
		}

		f.json(w, map[string]any{"data": data, "has_more": false})
	})
	mux.HandleFunc("GET /v1/checkout/sessions", func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()

		f.json(w, map[string]any{"data": f.openSessions, "has_more": false})
	})
	mux.HandleFunc("POST /v1/checkout/sessions/{id}/expire", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.expired = append(f.expired, r.PathValue("id"))
		f.mu.Unlock()

		f.json(w, map[string]any{"id": r.PathValue("id"), "status": "expired"})
	})
	mux.HandleFunc("GET /v1/subscriptions/{id}", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.subReads++
		sub, ok := f.subs[r.PathValue("id")]
		f.mu.Unlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			f.json(w, map[string]any{
				"error": map[string]any{"type": "invalid_request_error", "code": "resource_missing"},
			})

			return
		}

		f.json(w, sub)
	})

	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+stripeKey || r.Header.Get("Stripe-Version") != stripe.APIVersion {
			t.Errorf("%s %s: Authorization or Stripe-Version missing", r.Method, r.URL.Path)
		}

		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(f.srv.Close)

	return f
}

func (f *fakeStripe) json(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		f.t.Error(err)
	}
}

// stats returns what the fake API received so far.
func (f *fakeStripe) stats() (customers, subReads int, checkouts, portals []url.Values) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.customers, f.subReads, slices.Clone(f.checkouts), slices.Clone(f.portals)
}

// setSubscription makes the fake API serve the test subscription.
func (f *fakeStripe) setSubscription(status, price string) {
	f.addSubscription(testSubscription, status, price, 1_700_000_000)
}

// addSubscription makes the fake API serve subscription id of the test customer, created at the unix time created.
func (f *fakeStripe) addSubscription(id, status, price string, created int64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.subs[id] = map[string]any{
		"id": id, "customer": testCustomer, "status": status, "cancel_at_period_end": false, "created": created,
		"items": map[string]any{"data": []any{map[string]any{
			"id": "si_1", "price": map[string]any{"id": price}, "quantity": 1,
			"current_period_end": time.Now().Add(30 * 24 * time.Hour).Unix(),
		}}},
	}
}

// billingApp boots the app against f, with /test/subscribed and /test/business
// protected by the billing middleware.
func billingApp(t *testing.T, f *fakeStripe) *vukemtest.App {
	t.Helper()

	return vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{
			"STRIPE_KEY": stripeKey, "STRIPE_WEBHOOK_SECRET": webhookSecret, "STRIPE_API_BASE": f.srv.URL,
			"STRIPE_PRICE_PRO": "price_pro", "STRIPE_PRICE_BUSINESS": "price_business",
		},
		AppOptions: []vukem.Option{vukem.WithRoutes(protectedRoutes)},
	})
}

func protectedRoutes(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	b, err := billing.New(app)
	if err != nil {
		return err
	}

	ok := func(c *web.Ctx) error { return c.String(http.StatusOK, "members only") }
	r.Get("/test/subscribed", ok).Use(auth.Required(acc.Web), b.Subscribed())
	r.Get("/test/business", ok).Use(auth.Required(acc.Web), b.OnPlan("business"))

	return nil
}

func newUser(t *testing.T, app *vukemtest.App, email string) models.User {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	u := models.User{Name: "Ana Silva", Email: email, Password: "not used"}
	u.EmailVerifiedAt.V, u.EmailVerifiedAt.Valid = time.Now().UTC(), true

	if err := orm.Create(t.Context(), db, &u); err != nil {
		t.Fatal(err)
	}

	return u
}

// setCustomer links u to the test Stripe customer.
func setCustomer(t *testing.T, app *vukemtest.App, u models.User) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := orm.QueryOf[models.Billable](db).Where("id", u.ID).
		Update(t.Context(), orm.Set{"stripe_customer_id": testCustomer}); err != nil {
		t.Fatal(err)
	}
}

// postWebhook sends a signed Stripe event to the app.
func postWebhook(t *testing.T, app *vukemtest.App, id, typ string, object map[string]any) *vukemtest.Response {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"id": id, "type": typ, "created": time.Now().Unix(), "data": map[string]any{"object": object},
	})
	if err != nil {
		t.Fatal(err)
	}

	return app.Client(t).
		WithHeader("Stripe-Signature", stripe.SignForTest(payload, webhookSecret, time.Now())).
		Do(http.MethodPost, billing.WebhookPath, bytes.NewReader(payload), "application/json")
}

func completedCheckout(u models.User, status string) map[string]any {
	return map[string]any{
		"id": "cs_test_1", "mode": "subscription", "payment_status": status, "customer": "cus_1",
		"subscription": "sub_1", "client_reference_id": u.AuthID(),
	}
}

func TestBillingCheckout(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)

	app.Client(t).PostForm("/billing/checkout/business", nil).AssertRedirect("/login")

	u := newUser(t, app, "checkout@example.com")
	c := app.Client(t).ActingAs("web", u)

	c.PostForm("/billing/checkout/business", nil).
		AssertStatus(http.StatusSeeOther).
		AssertRedirect("https://checkout.stripe.com/c/pay/cs_test_1")
	vukemtest.AssertDatabaseHas(t, app, "users", map[string]any{"id": u.ID, "stripe_customer_id": "cus_1"})

	// The customer is created once.
	c.PostForm("/billing/checkout/pro", nil).AssertStatus(http.StatusSeeOther)

	customers, _, checkouts, _ := f.stats()
	if customers != 1 || len(checkouts) != 2 {
		t.Fatalf("customers = %d, checkouts = %d", customers, len(checkouts))
	}

	form := checkouts[0]
	for key, want := range map[string]string{
		"mode": "subscription", "customer": "cus_1", "client_reference_id": u.AuthID(),
		"line_items[0][price]": "price_business", "line_items[0][quantity]": "1",
	} {
		if got := form.Get(key); got != want {
			t.Errorf("checkout %s = %q, want %q", key, got, want)
		}
	}

	for key := range form {
		if strings.HasPrefix(key, "payment_method_types") {
			t.Errorf("checkout sent %s: payment methods come from the Dashboard", key)
		}
	}

	// Coming back from Checkout grants nothing: only the webhook does.
	c.Get("/billing?checkout=success").AssertOK()
	c.Get("/test/subscribed").AssertRedirect("/billing")

	c.PostForm("/billing/checkout/gold", nil).AssertNotFound()
}

func TestStripeWebhookSyncsSubscription(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "sync@example.com")
	setCustomer(t, app, u)
	f.setSubscription("active", "price_business")

	// A delayed payment method completes unpaid: no access yet.
	postWebhook(t, app, "evt_1", "checkout.session.completed", completedCheckout(u, "unpaid")).AssertOK()
	vukemtest.AssertDatabaseCount(t, app, "subscriptions", 0)

	postWebhook(t, app, "evt_2", "checkout.session.async_payment_succeeded", completedCheckout(u, "paid")).AssertOK()
	vukemtest.AssertDatabaseHas(t, app, "subscriptions", map[string]any{
		"user_id": u.ID, "stripe_subscription_id": "sub_1", "status": "active", "price_id": "price_business",
	})

	app.Client(t).ActingAs("web", u).Get("/test/business").AssertOK().AssertSee("members only")

	// A failed renewal makes it past_due; cancelling ends access.
	f.setSubscription("past_due", "price_business")
	postWebhook(t, app, "evt_3", "invoice.payment_failed", map[string]any{
		"id": "in_1", "parent": map[string]any{"subscription_details": map[string]any{"subscription": "sub_1"}},
	}).AssertOK()
	vukemtest.AssertDatabaseHas(t, app, "subscriptions", map[string]any{
		"stripe_subscription_id": "sub_1", "status": "past_due",
	})

	f.setSubscription("canceled", "price_business")
	postWebhook(t, app, "evt_4", "customer.subscription.deleted", map[string]any{
		"id": "sub_1", "customer": "cus_1",
	}).AssertOK()
	vukemtest.AssertDatabaseHas(t, app, "subscriptions", map[string]any{
		"stripe_subscription_id": "sub_1", "status": "canceled",
	})
	vukemtest.AssertDatabaseCount(t, app, "subscriptions", 1)

	app.Client(t).ActingAs("web", u).Get("/test/subscribed").AssertRedirect("/billing")
}

func TestStripeWebhookLinksCheckoutCustomer(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "link@example.com")
	f.setSubscription("trialing", "price_pro")

	postWebhook(t, app, "evt_1", "checkout.session.completed", completedCheckout(u, "no_payment_required")).AssertOK()
	vukemtest.AssertDatabaseHas(t, app, "users", map[string]any{"id": u.ID, "stripe_customer_id": "cus_1"})
	vukemtest.AssertDatabaseHas(t, app, "subscriptions", map[string]any{"user_id": u.ID, "status": "trialing"})

	// A session naming another user cannot take this customer over.
	other := newUser(t, app, "other@example.com")
	session := completedCheckout(other, "paid")
	postWebhook(t, app, "evt_2", "checkout.session.completed", session).AssertOK()
	vukemtest.AssertDatabaseHas(t, app, "users", map[string]any{"id": other.ID, "stripe_customer_id": nil})
}

func TestStripeWebhookProcessedOnce(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "once@example.com")
	setCustomer(t, app, u)
	f.setSubscription("active", "price_pro")

	event := map[string]any{"id": "sub_1", "customer": "cus_1"}
	postWebhook(t, app, "evt_dup", "customer.subscription.created", event).AssertOK()
	postWebhook(t, app, "evt_dup", "customer.subscription.created", event).AssertOK()

	if _, reads, _, _ := f.stats(); reads != 1 {
		t.Fatalf("the duplicate delivery was processed again (%d subscription reads)", reads)
	}

	vukemtest.AssertDatabaseCount(t, app, "subscriptions", 1)
}

// A08 (SDD 13): a replayed event id is dropped by contrib/stripe's cache-backed
// dedup, and a new event about the same Stripe object converges on one row.
func TestSecurity_A08_StripeReplayIgnored(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "replay@example.com")
	setCustomer(t, app, u)
	f.setSubscription("active", "price_pro")

	updated := map[string]any{"id": "sub_1", "customer": "cus_1"}
	paid := map[string]any{
		"id": "in_1", "parent": map[string]any{"subscription_details": map[string]any{"subscription": "sub_1"}},
	}

	for i, step := range []struct {
		id, typ string
		object  map[string]any
		reads   int
	}{
		{"evt_upd", "customer.subscription.updated", updated, 1},
		{"evt_upd", "customer.subscription.updated", updated, 1},
		{"evt_paid", "invoice.paid", paid, 2},
		{"evt_paid", "invoice.paid", paid, 2},
		{"evt_paid_2", "invoice.paid", paid, 3},
	} {
		postWebhook(t, app, step.id, step.typ, step.object).AssertOK()

		if _, reads, _, _ := f.stats(); reads != step.reads {
			t.Fatalf("step %d (%s %s): %d subscription reads, want %d", i, step.id, step.typ, reads, step.reads)
		}
	}

	vukemtest.AssertDatabaseCount(t, app, "subscriptions", 1)
	vukemtest.AssertDatabaseHas(t, app, "subscriptions", map[string]any{
		"user_id": u.ID, "stripe_subscription_id": "sub_1", "status": "active", "price_id": "price_pro",
	})
}

func TestStripeWebhookRejectsBadSignature(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "forged@example.com")
	setCustomer(t, app, u)
	f.setSubscription("active", "price_business")

	payload := []byte(`{"id":"evt_forged","type":"customer.subscription.created","data":{"object":{"id":"sub_1"}}}`)

	for name, sig := range map[string]string{
		"none":       "",
		"other key":  stripe.SignForTest(payload, "whsec_attacker", time.Now()),
		"replayed":   stripe.SignForTest(payload, webhookSecret, time.Now().Add(-time.Hour)),
		"not a sign": "t=1,v1=00",
	} {
		app.Client(t).WithHeader("Stripe-Signature", sig).
			Do(http.MethodPost, billing.WebhookPath, bytes.NewReader(payload), "application/json").
			AssertStatus(http.StatusBadRequest)

		if _, reads, _, _ := f.stats(); reads != 0 {
			t.Fatalf("%s: a forged event was processed", name)
		}
	}

	vukemtest.AssertDatabaseCount(t, app, "subscriptions", 0)
}

func TestBillingMiddleware(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)

	app.Client(t).Get("/test/subscribed").AssertRedirect("/login")

	u := newUser(t, app, "member@example.com")
	c := app.Client(t).ActingAs("web", u)

	c.Get("/test/subscribed").AssertRedirect("/billing")
	c.AcceptJSON().Get("/test/subscribed").AssertStatus(http.StatusPaymentRequired)

	setCustomer(t, app, u)
	f.setSubscription("active", "price_pro")
	postWebhook(t, app, "evt_1", "customer.subscription.created", map[string]any{"id": "sub_1"}).AssertOK()

	member := app.Client(t).ActingAs("web", u)
	member.Get("/test/subscribed").AssertOK().AssertSee("members only")
	member.Get("/test/business").AssertRedirect("/billing") // pro is not business

	// Subscribers manage their plan in the portal, not with a new checkout.
	member.PostForm("/billing/checkout/business", nil).AssertRedirect("/billing")

	if _, _, checkouts, _ := f.stats(); len(checkouts) != 0 {
		t.Fatalf("a subscriber started a second checkout")
	}
}

func TestBillingPortal(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "portal@example.com")
	c := app.Client(t).ActingAs("web", u)

	// No customer yet: back to the plans.
	c.PostForm("/billing/portal", nil).AssertRedirect("/billing")

	setCustomer(t, app, u)
	c.PostForm("/billing/portal", nil).
		AssertStatus(http.StatusSeeOther).
		AssertRedirect("https://billing.stripe.com/p/session/test_1")

	_, _, _, portals := f.stats()
	if len(portals) != 1 || portals[0].Get("customer") != "cus_1" ||
		portals[0].Get("return_url") != app.Config.URL+"/billing" {
		t.Fatalf("portal sessions = %v", portals)
	}
}

func TestBillingNotConfigured(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	u := newUser(t, app, "nokey@example.com")

	app.Client(t).ActingAs("web", u).PostForm("/billing/checkout/business", nil).AssertNotFound() // no price
	app.Client(t).ActingAs("web", u).PostForm("/billing/portal", nil).AssertStatus(http.StatusServiceUnavailable)
	app.Client(t).Do(http.MethodPost, billing.WebhookPath, bytes.NewReader([]byte(`{}`)), "application/json").
		AssertStatus(http.StatusServiceUnavailable)
}

func TestBillingCheckoutCurrency(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	c := app.Client(t).ActingAs("web", newUser(t, app, "currency@example.com"))

	c.PostForm("/billing/checkout/business", url.Values{"currency": {"BRL"}}).AssertStatus(http.StatusSeeOther)
	c.PostForm("/billing/checkout/pro", nil).AssertStatus(http.StatusSeeOther)
	c.PostForm("/billing/checkout/pro", url.Values{"currency": {"gbp"}}).AssertStatus(http.StatusUnprocessableEntity)

	_, _, checkouts, _ := f.stats()
	if len(checkouts) != 2 {
		t.Fatalf("checkouts = %d, want 2 (unsupported currency must not reach Stripe)", len(checkouts))
	}

	for i, want := range []struct{ price, currency string }{{"price_business", "brl"}, {"price_pro", "usd"}} {
		if got := checkouts[i].Get("currency"); got != want.currency {
			t.Errorf("checkout %d currency = %q, want %q", i, got, want.currency)
		}

		if got := checkouts[i].Get("line_items[0][price]"); got != want.price {
			t.Errorf("checkout %d price = %q, want %q", i, got, want.price)
		}
	}
}

// TestStripeWebhookSupersedesSubscription: two live subscriptions (two Checkout tabs) keep the newest by Stripe
// created time and cancel the other at Stripe, whatever order the webhooks arrive in.
func TestStripeWebhookSupersedesSubscription(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "twice@example.com")
	setCustomer(t, app, u)

	sync := func(evt, id string) {
		t.Helper()
		postWebhook(t, app, evt, "customer.subscription.updated", map[string]any{"id": id, "customer": "cus_1"}).
			AssertOK()
	}

	f.addSubscription("sub_1", "active", "price_pro", 1_700_000_000)
	sync("evt_s1", "sub_1")

	// A newer subscription supersedes sub_1.
	f.addSubscription("sub_2", "active", "price_business", 1_700_000_100)
	sync("evt_s2", "sub_2")

	// An older one arriving late is cancelled itself and does not unseat sub_2.
	f.addSubscription("sub_0", "active", "price_pro", 1_600_000_000)
	sync("evt_s0", "sub_0")

	// sub_1's own late event (now canceled at Stripe) changes nothing either.
	sync("evt_s1b", "sub_1")

	f.mu.Lock()
	cancels := slices.Clone(f.cancels)
	f.mu.Unlock()
	if slices.Sort(cancels); !slices.Equal(cancels, []string{"sub_0", "sub_1"}) {
		t.Fatalf("cancelled at Stripe = %v, want [sub_0 sub_1]", cancels)
	}

	for id, state := range map[string]string{"sub_2": "active", "sub_1": "canceled", "sub_0": "canceled"} {
		vukemtest.AssertDatabaseHas(
			t,
			app,
			"subscriptions",
			map[string]any{"stripe_subscription_id": id, "state": state},
		)
	}
	app.Client(t).ActingAs("web", u).Get("/test/business").AssertOK()
}

// TestBillingCheckoutExpiresOpenSessions: a new Checkout expires the customer's other open subscription sessions,
// so a second tab cannot start a second subscription.
func TestBillingCheckoutExpiresOpenSessions(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "tabs@example.com")
	setCustomer(t, app, u)
	f.openSessions = []map[string]any{
		{"id": "cs_old", "mode": "subscription", "status": "open"},
		{"id": "cs_pay", "mode": "payment", "status": "open"},
	}

	app.Client(t).ActingAs("web", u).PostForm("/billing/checkout/pro", nil).AssertStatus(http.StatusSeeOther)

	f.mu.Lock()
	defer f.mu.Unlock()
	if !slices.Equal(f.expired, []string{"cs_old"}) {
		t.Fatalf("expired = %v, want [cs_old]", f.expired)
	}
}
