package legacyimport_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

func TestIgnoreNoLoginDropsOnlyThatCheck(t *testing.T) {
	t.Parallel()

	only := legacyimport.PreflightResult{Failures: map[string]int64{"no_login_method": 10000}}
	_, err := legacyimport.IgnoreNoLogin(only, &legacyimport.PreflightError{Failures: only.Failures})
	if err != nil {
		t.Fatal(err)
	}

	both := legacyimport.PreflightResult{Failures: map[string]int64{
		"no_login_method": 1, "duplicate_emails": 2,
	}}
	_, err = legacyimport.IgnoreNoLogin(both, &legacyimport.PreflightError{Failures: both.Failures})
	if err == nil || !strings.Contains(err.Error(), "duplicate_emails=2") ||
		strings.Contains(err.Error(), "no_login_method") {
		t.Fatalf("err = %v", err)
	}

	_, err = legacyimport.IgnoreNoLogin(legacyimport.PreflightResult{}, errors.New("db down"))
	if err == nil || err.Error() != "db down" {
		t.Fatal(err)
	}
}
