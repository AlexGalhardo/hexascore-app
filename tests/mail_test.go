package tests

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestMailLocaleFromPayload: mail queued outside a request (no locale in the context) renders in the locale
// saved on the user, subject and body.
func TestMailLocaleFromPayload(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct{ email, locale, subject, body string }{
		{"pt@example.com", "pt-BR", "Sua conta está agendada para exclusão", "Olá"},
		{"es@example.com", "es", "Tu cuenta está programada para eliminarse", "Hola"},
		{"en@example.com", "en", "Your account is scheduled for deletion", "Hello"},
	}
	for _, c := range cases {
		register(t, app, app.Client(t), c.email)
		if _, err := db.ExecContext(t.Context(), `UPDATE users SET locale = `+db.Dialect().Placeholder(1)+
			` WHERE email = `+db.Dialect().Placeholder(2), c.locale, c.email); err != nil {
			t.Fatal(err)
		}
		var id int64
		if err := db.QueryRowContext(t.Context(), `SELECT id FROM users WHERE email = `+db.Dialect().Placeholder(1), c.email).
			Scan(&id); err != nil {
			t.Fatal(err)
		}

		if _, err := acc.ScheduleDeletion(t.Context(), id); err != nil {
			t.Fatal(err)
		}

		var found bool
		for _, m := range app.Mails().SentTo(c.email) {
			if m.Subject == c.subject {
				found = true
				if !strings.Contains(m.HTML, c.body) || !strings.Contains(m.HTML, `lang="`+c.locale+`"`) {
					t.Errorf("%s: body not in %s: %s", c.email, c.locale, m.HTML)
				}
			}
		}
		if !found {
			t.Errorf("%s: no mail with subject %q", c.email, c.subject)
		}
	}
}

// TestAutomaticMailFromAndReplyTo: a real verification mail leaves from noreply@ with Reply-To support@ (Q-A17).
func TestAutomaticMailFromAndReplyTo(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New)
	register(t, app, app.Client(t), "sender@example.com")

	sent := app.Mails().SentTo("sender@example.com")
	if len(sent) == 0 {
		t.Fatal("no mail sent")
	}
	for _, m := range sent {
		if m.From != "HexaScore <noreply@hexascore.app>" && !strings.Contains(m.From, "noreply@hexascore.app") {
			t.Errorf("From = %q", m.From)
		}
		if len(m.ReplyTo) != 1 || !strings.Contains(m.ReplyTo[0], "support@hexascore.app") {
			t.Errorf("ReplyTo = %v", m.ReplyTo)
		}
	}
}
