package problems_test

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/http/problems"
)

func TestForStatusSkipsSpecificCodes(t *testing.T) {
	t.Parallel()

	cases := map[int]string{
		http.StatusForbidden:       "plan-required",
		http.StatusTooManyRequests: "rate-limited",
		http.StatusConflict:        "",
		http.StatusLocked:          "",
	}
	for status, want := range cases {
		if got := problems.ForStatus(status); got != want {
			t.Errorf("ForStatus(%d) = %q, want %q", status, got, want)
		}
	}

	for _, code := range []string{"suggestion-duplicate", "field-locked", "community-submissions-closed"} {
		if !problems.Known(code) {
			t.Errorf("%s not in catalogue", code)
		}
	}
}
