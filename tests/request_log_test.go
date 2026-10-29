//go:build integration

package tests

import (
	"bytes"
	"strings"
	"testing"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// The privacy handler must wrap the logger vukem's request Logger uses
// (bootstrap wires it before routes.Web installs the default stack).
func TestLog_BootstrapWiresPrivacyHandler(t *testing.T) {
	var buf bytes.Buffer

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:        map[string]string{"DATABASE_URL": pgtest.Database(t, ""), "LOG_FORMAT": "json"},
		AppOptions: []vukem.Option{vukem.WithLogOutput(&buf)},
	})

	app.Client(t).Get("/this-route-does-not-exist?q=secret+words")

	var line string
	for l := range strings.SplitSeq(buf.String(), "\n") {
		if strings.Contains(l, `"msg":"request"`) {
			line = l
		}
	}

	if line == "" {
		t.Fatalf("no request line:\n%s", buf.String())
	}

	for _, want := range []string{`"route":"`, `"duration_ms":`, `"request_id":"`} {
		if !strings.Contains(line, want) {
			t.Errorf("request line lacks %s: %s", want, line)
		}
	}

	if strings.Contains(buf.String(), "secret words") || strings.Contains(line, `"duration":`) {
		t.Errorf("request line not reshaped: %s", line)
	}
}
