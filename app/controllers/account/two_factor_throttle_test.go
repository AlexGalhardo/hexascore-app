package account_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/controllers/account"
)

type pendingSession struct{ web.Session }

func (pendingSession) Get(key string) any {
	if key == "two_factor.user" {
		return "42"
	}

	return nil
}

func TestSecurity_A07_TwoFactorAccountKeyIgnoresIP(t *testing.T) {
	t.Parallel()

	req := func(method, remote string) string {
		r := httptest.NewRequest(method, "/2fa", nil)
		r.RemoteAddr = remote
		r = r.WithContext(web.WithSession(r.Context(), pendingSession{}))

		return account.TwoFactorAccountKey(r)
	}

	a, b := req(http.MethodPost, "203.0.113.1:4000"), req(http.MethodPost, "198.51.100.9:4000")
	if a == "" || a != b {
		t.Fatalf("account bucket = %q and %q, want one non-empty key whatever the address", a, b)
	}

	if req(http.MethodGet, "203.0.113.1:4000") != "" {
		t.Fatal("GET must not consume the bucket")
	}

	if got := account.TwoFactorAccountKey(httptest.NewRequest(http.MethodPost, "/2fa", nil)); got != "" {
		t.Fatalf("no pending login gave key %q", got)
	}
}
