package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func TestNicknameRequired(t *testing.T) {
	signedIn := func(*web.Ctx) (models.User, error) { return models.User{}, nil }
	signedOut := func(*web.Ctx) (models.User, error) { return models.User{}, errors.New("guest") }
	nick := func(has bool) func(*web.Ctx, int64) (bool, error) {
		return func(*web.Ctx, int64) (bool, error) { return has, nil }
	}

	cases := []struct {
		name       string
		user       func(*web.Ctx) (models.User, error)
		has        bool
		method     string
		header     map[string]string
		wantStatus int
		wantTo     string
		wantCode   string
	}{
		{"read passes", signedIn, false, http.MethodGet, nil, http.StatusNoContent, "", ""},
		{"guest write passes", signedOut, false, http.MethodPost, nil, http.StatusNoContent, "", ""},
		{"write with nickname passes", signedIn, true, http.MethodPost, nil, http.StatusNoContent, "", ""},
		{"web write redirects", signedIn, false, http.MethodPost, nil, http.StatusSeeOther, "/onboarding", ""},
		{"inertia write redirects", signedIn, false, http.MethodPost, map[string]string{
			"X-Inertia": "true", "Accept": "application/json",
		}, http.StatusSeeOther, "/onboarding", ""},
		{"json write is refused", signedIn, false, http.MethodDelete, map[string]string{
			"Accept": "application/json",
		}, http.StatusForbidden, "", "nickname-required"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := middleware.NicknameRequired(tc.user, nick(tc.has), "https://example.test/problems")(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
			)

			r := httptest.NewRequest(tc.method, "/x", nil)
			for k, v := range tc.header {
				r.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantStatus)
			}

			if got := w.Header().Get("Location"); got != tc.wantTo {
				t.Errorf("Location = %q, want %q", got, tc.wantTo)
			}

			if tc.wantCode != "" && !strings.Contains(w.Body.String(), `"code":"`+tc.wantCode+`"`) {
				t.Errorf("body = %s, want code %s", w.Body.String(), tc.wantCode)
			}
		})
	}
}
