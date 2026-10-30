package limits

import (
	"net/http"
	"strings"

	"github.com/AlexGalhardo/vukem/web"
)

// Body limits of SDD 13 §7.3. Suggestion JSON and cover uploads ship with their P4.A routes; the
// ClientErrorsBody and ClientMetricsBody cap POST /app/v1/client-errors and
// POST /app/v1/client-metrics. BodyFor special-cases those paths because the
// server-level limit runs before routing.
const (
	FormBody          int64 = 64 << 10
	SuggestionBody    int64 = 256 << 10
	CoverBody         int64 = 5 << 20
	StripeWebhookBody int64 = 512 << 10
	ClientErrorsBody  int64 = 16 << 10
	CSPReportsBody    int64 = 16 << 10
	ClientMetricsBody int64 = 4 << 10
	DefaultBody       int64 = 1 << 20
)

// stripeWebhookPath is billing.WebhookPath (billing imports this package); the webhook limit test
// fails if the two drift.
const stripeWebhookPath = "/stripe/webhook"

// BodyFor is the limit of a request path: the JSON stacks (/api/v1, /app/v1) take DefaultBody, the
// Stripe webhook StripeWebhookBody, and every other route (web and admin) is a form.
func BodyFor(path string) int64 {
	if path == stripeWebhookPath {
		return StripeWebhookBody
	}

	if path == "/csp-reports" {
		return CSPReportsBody
	}

	if path == "/app/v1/client-errors" {
		return ClientErrorsBody
	}

	if path == "/app/v1/client-metrics" {
		return ClientMetricsBody
	}

	if strings.HasPrefix(path, "/api/v1/") || strings.HasPrefix(path, "/app/v1/") {
		return DefaultBody
	}

	return FormBody
}

// Body applies BodyFor; as router middleware it replaces vukem's 10 MiB router default. A route with a tighter or larger limit
// (an upload's Rules.MaxSize, a later BodyLimit) replaces it again, as vukem's BodyLimit documents.
func Body(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, web.LimitBody(w, r, BodyFor(r.URL.Path)))
	})
}
