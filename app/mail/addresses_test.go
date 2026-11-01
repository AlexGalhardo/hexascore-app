package mail

import (
	"testing"

	vmail "github.com/AlexGalhardo/vukem/mail"
)

func TestAutomaticMailSenderAndReplyTo(t *testing.T) {
	Configure("", "")
	cases := map[string]vmail.Envelope{
		"verify":    VerifyEmail{}.Envelope(),
		"reset":     ResetPassword{}.Envelope(),
		"exists":    AccountExists{}.Envelope(),
		"export":    ExportReady{}.Envelope(),
		"community": SuggestionNotice{}.Envelope(),
	}
	for name, e := range cases {
		if e.From != "HexaScore <noreply@hexascore.app>" {
			t.Errorf("%s: From = %q", name, e.From)
		}
		if len(e.ReplyTo) != 1 || e.ReplyTo[0] != "support@hexascore.app" {
			t.Errorf("%s: ReplyTo = %v", name, e.ReplyTo)
		}
	}
}

func TestConfigureOverrides(t *testing.T) {
	t.Cleanup(func() { Configure("", "") })
	Configure("A <a@example.com>", "s@example.com")
	e := VerifyEmail{}.Envelope()
	if e.From != "A <a@example.com>" || e.ReplyTo[0] != "s@example.com" {
		t.Fatalf("got %+v", e)
	}
}

// The contact notification keeps the visitor as Reply-To so support answers them.
func TestContactKeepsVisitorReplyTo(t *testing.T) {
	Configure("", "")
	e := ContactMessage{Inbox: "support@hexascore.app", Email: "v@example.com"}.Envelope()
	if e.From != "HexaScore <noreply@hexascore.app>" || e.ReplyTo[0] != "v@example.com" {
		t.Fatalf("got %+v", e)
	}
}
