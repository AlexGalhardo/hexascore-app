//go:build integration

package tests

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/webhooks"
)

func TestAPIWebhooks(t *testing.T) {
	r, client := webhookClient(t, "hooks-pro@example.com", "pro")

	body := map[string]any{
		"url": "https://1.1.1.1/hook", "events": []string{"item.updated"}, "itemTypes": []string{"games"},
	}
	before := quotaUsed(t, client)
	created := client.PostJSON("/api/v1/webhooks", body).AssertCreated().AssertMatchesOpenAPI()
	var one struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}
	created.JSON(&one)
	if one.Secret == "" || one.ID == "" {
		t.Fatalf("create body = %#v", one)
	}

	after := quotaUsed(t, client)
	if after != before+1 {
		t.Fatalf("quota used %d → %d, want +1 (the usage read only)", before, after)
	}

	second := client.PostJSON("/api/v1/webhooks", map[string]any{
		"url": "https://1.0.0.1/hook", "events": []string{"item.created", "item.deleted"}, "itemTypes": []string{"movies"},
	}).AssertCreated()
	var two struct {
		ID string `json:"id"`
	}
	second.JSON(&two)
	client.PostJSON("/api/v1/webhooks", map[string]any{
		"url": "https://8.8.8.8/hook", "events": []string{"item.updated"}, "itemTypes": []string{"books"},
	}).AssertUnprocessable().AssertJSONPath("code", "webhook-limit")

	listed := client.GetJSON("/api/v1/webhooks").AssertOK().AssertMatchesOpenAPI().AssertJSONCount("data", 2)
	listed.AssertDontSee(one.Secret).AssertDontSee("whsec_")
	listed.AssertJSONPath("data.0.url", "https://1.1.1.1/hook")

	_, other := webhookClient(t, "hooks-other@example.com", "pro")
	other.Delete("/api/v1/webhooks/" + two.ID).AssertNotFound()
	client.GetJSON("/api/v1/webhooks").AssertOK().AssertJSONCount("data", 2)

	client.Delete("/api/v1/webhooks/" + one.ID).AssertNoContent()
	client.GetJSON("/api/v1/webhooks").AssertOK().AssertJSONCount("data", 1)

	_, free := webhookClient(t, "hooks-free@example.com", "")
	free.PostJSON("/api/v1/webhooks", body).AssertForbidden().AssertJSONPath("code", "plan-required")

	_, biz := webhookClient(t, "hooks-biz@example.com", "business")
	biz.PostJSON("/api/v1/webhooks", body).AssertCreated()
	biz.PostJSON("/api/v1/webhooks", map[string]any{
		"url": "https://1.0.0.1/hook", "events": []string{"item.updated"}, "itemTypes": []string{"series"},
	}).AssertCreated()
	biz.PostJSON("/api/v1/webhooks", map[string]any{
		"url": "https://8.8.8.8/hook", "events": []string{"item.updated"}, "itemTypes": []string{"games"},
	}).AssertCreated()

	_ = r
}

func TestSecurity_A10_WebhookSSRF(t *testing.T) {
	_, client := webhookClient(t, "hooks-ssrf@example.com", "pro")
	for _, raw := range []string{
		"https://127.0.0.1/hook",
		"https://10.1.2.3/hook",
		"http://1.1.1.1/hook",
	} {
		client.PostJSON("/api/v1/webhooks", map[string]any{
			"url": raw, "events": []string{"item.updated"}, "itemTypes": []string{"games"},
		}).AssertUnprocessable().AssertJSONPath("code", "webhook-url")
	}
}

func TestSecurity_A02_WebhookSecretEncrypted(t *testing.T) {
	r, client := webhookClient(t, "hooks-secret@example.com", "pro")
	res := client.PostJSON("/api/v1/webhooks", map[string]any{
		"url": "https://1.1.1.1/hook", "events": []string{"item.updated"}, "itemTypes": []string{"games"},
	}).AssertCreated()
	var body struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}
	res.JSON(&body)

	var stored string
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT secret_enc FROM webhook_endpoints WHERE id = $1`, body.ID).Scan(&stored); err != nil {
		t.Fatal(err)
	}

	if stored == body.Secret || stored == "" {
		t.Fatal("secret column holds the plaintext")
	}

	if len(stored) >= 6 && stored[:6] == "whsec_" {
		t.Fatal("secret column starts with whsec_")
	}

	plain, err := r.app.App.Crypt.Open(stored, webhooks.SecretAD)
	if err != nil {
		t.Fatal(err)
	}

	if string(plain) != body.Secret {
		t.Fatal("opened secret does not match the one-time body")
	}
}

func webhookClient(t *testing.T, email, plan string) (*keyRig, *vukemtest.Client) {
	t.Helper()

	r := newKeyRigWith(t,
		vukem.WithRateLimiters(func(_ string, _ int, window time.Duration) ratelimit.Limiter {
			return ratelimit.NewMemory(1000, window)
		}),
	)
	token := r.create(r.signIn(email), "ci")
	if plan != "" {
		setPlan(t, r, email, plan)
	}

	return r, r.app.Client(t).WithToken(token)
}

func setPlan(t *testing.T, r *keyRig, email, plan string) {
	t.Helper()

	var id int64
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT id FROM users WHERE email = $1`, email).Scan(&id); err != nil {
		t.Fatal(err)
	}

	_, err := r.acc.DB.ExecContext(t.Context(), `INSERT INTO subscriptions (
  user_id, stripe_subscription_id, status, price_id, plan, provider, state,
  current_period_end, cancel_at_period_end, created_at, updated_at)
VALUES ($1, $2, 'active', 'price_test', $3, 'stripe', 'active', now() + interval '30 days', false, now(), now())`,
		id, "sub_"+email, plan)
	if err != nil {
		t.Fatal(err)
	}
}

func quotaUsed(t *testing.T, client *vukemtest.Client) int {
	t.Helper()

	var body struct {
		Quota struct {
			Used int `json:"used"`
		} `json:"quota"`
	}
	client.GetJSON("/api/v1/me/usage").AssertOK().JSON(&body)

	return body.Quota.Used
}
