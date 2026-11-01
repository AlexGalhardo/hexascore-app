package config_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/config"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestMailAddressDefaults(t *testing.T) {
	get := env(nil)
	if got := config.MailFrom(get); got != "HexaScore <noreply@hexascore.app>" {
		t.Fatalf("MailFrom = %q", got)
	}
	if got := config.SupportAddress(get); got != "support@hexascore.app" {
		t.Fatalf("SupportAddress = %q", got)
	}
	if got := config.ContactInbox(get); got != "support@hexascore.app" {
		t.Fatalf("ContactInbox = %q, want the support address", got)
	}
}

func TestMailAddressOverrides(t *testing.T) {
	get := env(map[string]string{
		"MAIL_FROM": "X <x@example.com>", "MAIL_SUPPORT_ADDRESS": "help@example.com",
	})
	if config.MailFrom(get) != "X <x@example.com>" || config.SupportAddress(get) != "help@example.com" {
		t.Fatal("overrides ignored")
	}
	if got := config.ContactInbox(get); got != "help@example.com" {
		t.Fatalf("ContactInbox follows the support address, got %q", got)
	}
	if got := config.ContactInbox(
		env(map[string]string{"CONTACT_INBOX": "ops@example.com"}),
	); got != "ops@example.com" {
		t.Fatalf("explicit CONTACT_INBOX ignored: %q", got)
	}
}

// TestOnlyDocumentedAddresses: outside the config defaults, the only
// @hexascore.app mailbox in user-visible sources is privacy@ (D-39).
func TestOnlyDocumentedAddresses(t *testing.T) {
	re := regexp.MustCompile(`[A-Za-z0-9._-]+@hexascore\.app`)
	allowed := map[string]bool{
		"privacy@hexascore.app": true,
		"support@hexascore.app": true,
		"noreply@hexascore.app": true,
	}
	for _, dir := range []string{"../resources/lang", "../resources/js", "../apps/mobile/src", "../app", "../routes"} {
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			if strings.HasSuffix(p, "_test.go") {
				return nil
			}
			switch filepath.Ext(p) {
			case ".json", ".ts", ".tsx", ".go":
			default:
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			for _, m := range re.FindAllString(string(b), -1) {
				if !allowed[strings.ToLower(m)] {
					t.Errorf("%s: undocumented address %s", p, m)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
