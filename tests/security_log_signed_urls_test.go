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

// SDD 13 A09: a signed or temporary link is a working credential until it expires, so no log line may carry its
// signature, expiry or token: not the request line, not the error line, not a message that quotes the URL (mail
// bodies, wrapped client errors). The path and the status stay.
func TestSecurity_A09_SignedURLsNotLogged(t *testing.T) {
	var buf bytes.Buffer

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{
			"DATABASE_URL": pgtest.Database(t, ""), "LOG_FORMAT": "json", "LOG_LEVEL": "debug",
		},
		AppOptions: []vukem.Option{vukem.WithLogOutput(&buf)},
	})

	secrets := []string{
		"SIGEXPORT0001", "SIGOFFSITE001", "SIGVERIFY0001", "RESETTOKEN0001", "SIGTEXT000001",
		"AMZSIG0000001", "AMZCRED000001", "4102444800",
	}

	for _, path := range []string{
		"/exports/u1/library.zip?expires=4102444800&signature=SIGEXPORT0001",
		"/backups/offsite/postgres/monthly/2026-10.dump.age?expires=4102444800&signature=SIGOFFSITE001",
		"/verify-email/confirm/1?expires=4102444800&signature=SIGVERIFY0001",
		"/reset-password/RESETTOKEN0001",
		"/pt-br/reset-password/RESETTOKEN0001",
	} {
		app.Client(t).Get(path)
	}

	app.Log.Info("mail sent", "body",
		"Download: https://hexascore.app/exports/u1/x.zip?expires=4102444800&signature=SIGTEXT000001 (7 days)")
	app.Log.Warn("fetch failed", "err",
		`Get "https://bucket.s3/x?X-Amz-Credential=AMZCRED000001&X-Amz-Signature=AMZSIG0000001": timeout`)

	out := buf.String()
	for _, s := range secrets {
		if strings.Contains(out, s) {
			t.Errorf("log carries %q:\n%s", s, out)
		}
	}

	for _, want := range []string{
		`"path":"/exports/u1/library.zip"`, `"path":"/reset-password/[REDACTED]"`,
		`"path":"/pt-br/reset-password/[REDACTED]"`, `"status":`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("log lacks %s:\n%s", want, out)
		}
	}
}
