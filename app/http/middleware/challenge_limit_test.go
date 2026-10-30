package middleware_test

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
)

func TestGlobalLimitOverride(t *testing.T) {
	t.Parallel()

	n, err := middleware.GlobalLimitOverride("")
	if err != nil || n != 0 {
		t.Fatalf("empty = %d, %v; want 0, nil", n, err)
	}

	n, err = middleware.GlobalLimitOverride("100000")
	if err != nil || n != 100000 {
		t.Fatalf("100000 = %d, %v", n, err)
	}

	for _, raw := range []string{"0", "-3", "nope", "1.5"} {
		_, err := middleware.GlobalLimitOverride(raw)
		if err == nil || !strings.Contains(err.Error(), "CHALLENGE_GLOBAL_LIMIT") {
			t.Fatalf("%q: err = %v, want a refusal", raw, err)
		}
	}
}
