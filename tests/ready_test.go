package tests

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestReady_ChecksPresent pins FR-12-5: /up?checks=1 reports db, ssr, storage
// and scheduler. No Node server runs in tests, so ssr is the one failing check
// and the probe must still answer within one check timeout.
func TestReady_ChecksPresent(t *testing.T) {
	t.Parallel()

	if bootstrap.ReadyCheckTimeout != 2*time.Second {
		t.Fatalf("ReadyCheckTimeout = %s, want 2s (SDD 12 §7.1)", bootstrap.ReadyCheckTimeout)
	}

	app := vukemtest.New(t, bootstrap.New)

	start := time.Now()
	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	app.Client(t).Get("/up?checks=1").AssertStatus(503).JSON(&body)

	if took := time.Since(start); took > bootstrap.ReadyCheckTimeout+time.Second {
		t.Errorf("readiness took %s, want under %s", took, bootstrap.ReadyCheckTimeout+time.Second)
	}

	want := map[string]string{"db": "ok", "ssr": "fail", "storage": "ok", "scheduler": "ok"}
	if len(body.Checks) != len(want) {
		t.Errorf("checks = %v, want exactly %v", body.Checks, want)
	}

	for name, status := range want {
		if body.Checks[name] != status {
			t.Errorf("check %s = %q, want %q (all: %v)", name, body.Checks[name], status, body.Checks)
		}
	}

	app.Client(t).Get("/up").AssertOK()
}

// TestReady_SchedulerCheckOnlyInServe: with SCHEDULE_IN_SERVE=false the
// in-memory heartbeat never beats in this process, so the check is not added.
func TestReady_SchedulerCheckOnlyInServe(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"SCHEDULE_IN_SERVE": "false"}})

	var body struct {
		Checks map[string]string `json:"checks"`
	}
	app.Client(t).Get("/up?checks=1").JSON(&body)

	if _, ok := body.Checks["scheduler"]; ok {
		t.Errorf("checks = %v, want no scheduler check", body.Checks)
	}
}
