package publicapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/errs"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

func TestSearchErr_TimeoutIs503WithRetryAfter(t *testing.T) {
	t.Parallel()

	var he *errs.HTTPError
	if !errors.As(searchErr(fmt.Errorf("x: %w", search.ErrTimeout)), &he) {
		t.Fatal("not an HTTPError")
	}

	if he.Status != http.StatusServiceUnavailable || he.Headers.Get("Retry-After") != "5" {
		t.Fatalf("status %d headers %v", he.Status, he.Headers)
	}
}
