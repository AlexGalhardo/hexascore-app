package appapi

import (
	"cmp"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"
)

const (
	// ClientVersionHeader is the build the app is running (SDD 05 §6.4).
	ClientVersionHeader = "X-App-Version"
	// MinVersionHeader is the oldest build the server still accepts.
	MinVersionHeader = "X-App-Min-Version"
	// UpgradeCode is the problem code of a 426.
	UpgradeCode = "app-upgrade-required"
)

// MinVersion rejects /app/v1 calls whose X-App-Version is below
// APP_MIN_SUPPORTED_VERSION (default 0.0.0). DELETE /app/v1/tokens/current
// stays open so an old build can still log out. Other paths are unchanged.
func MinVersion(get func(string) string) func(http.Handler) http.Handler {
	supported := cmp.Or(strings.TrimSpace(get("APP_MIN_SUPPORTED_VERSION")), "0.0.0")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/app/v1/") && r.URL.Path != "/app/v1" {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set(MinVersionHeader, supported)

			if logoutExempt(r) || !versionBelow(r.Header.Get(ClientVersionHeader), supported) {
				next.ServeHTTP(w, r)
				return
			}

			web.Fail(w, r, errs.New(http.StatusUpgradeRequired, "This app build is no longer supported.").
				WithExtension("code", UpgradeCode).
				WithHeader(MinVersionHeader, supported))
		})
	}
}

func logoutExempt(r *http.Request) bool {
	return r.Method == http.MethodDelete && r.URL.Path == "/app/v1/tokens/current"
}

// versionBelow reports whether client is older than min. A missing or
// unparsable client version counts as 0.0.0. An unparsable minimum rejects
// every build, so a bad APP_MIN_SUPPORTED_VERSION fails closed.
func versionBelow(client, lowest string) bool {
	floor, ok := versionParts(lowest)
	if !ok {
		return true
	}

	got, ok := versionParts(client)
	if !ok {
		got = [3]int{}
	}

	for i := range floor {
		if got[i] != floor[i] {
			return got[i] < floor[i]
		}
	}

	return false
}

func versionParts(s string) ([3]int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return [3]int{}, false
	}

	bits := strings.Split(s, ".")
	if len(bits) == 0 || len(bits) > 3 {
		return [3]int{}, false
	}

	var out [3]int

	for i, b := range bits {
		b, _, _ = strings.Cut(b, "-")
		b, _, _ = strings.Cut(b, "+")

		n, err := strconv.Atoi(b)
		if err != nil || n < 0 {
			return [3]int{}, false
		}

		out[i] = n
	}

	return out, true
}
