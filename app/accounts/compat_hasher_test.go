package accounts_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/auth/hash"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/forms"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type betterAuthFixture struct {
	Label          string `json:"label"`
	Password       string `json:"password"`
	NFKCEquivalent string `json:"nfkc_equivalent"`
	Hash           string `json:"hash"`
}

func loadFixtures(t *testing.T) []betterAuthFixture {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/betterauth_hashes.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Cases []betterAuthFixture `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Cases) < 3 {
		t.Fatalf("got %d fixtures, want >= 3", len(doc.Cases))
	}
	return doc.Cases
}

func TestCompatHasherBetterAuthFixtures(t *testing.T) {
	h := accounts.NewCompatHasher(hash.Fake())
	for _, f := range loadFixtures(t) {
		t.Run(f.Label, func(t *testing.T) {
			if ok, err := h.Verify(f.Hash, f.Password); err != nil || !ok {
				t.Fatalf("Verify(fixture) = %v, %v; want true", ok, err)
			}
			if f.NFKCEquivalent != "" {
				if ok, _ := h.Verify(f.Hash, f.NFKCEquivalent); !ok {
					t.Fatal("NFKC-equivalent password must verify")
				}
			}
			if ok, _ := h.Verify(f.Hash, f.Password+"x"); ok {
				t.Fatal("wrong password verified")
			}
			if !h.NeedsRehash(f.Hash) {
				t.Fatal("Better Auth hash must need a rehash")
			}
		})
	}
	// A malformed legacy-looking value never verifies.
	if ok, _ := h.Verify(strings.Repeat("a", 32)+":zz", "x"); ok {
		t.Fatal("malformed hash verified")
	}
}

func TestSecurity_A07_LegacyScryptVerifiesAndRehashes(t *testing.T) {
	f := loadFixtures(t)[0]
	app := vukemtest.New(t, bootstrap.New)
	c := app.Client(t).Inertia("").WithHeader("Referer", "/login")
	const email = "legacy@example.com"

	c.PostJSON("/signup", map[string]any{
		"name": "Legacy User", "email": email, "password": "Temporary pass 1!",
		"password_confirmation": "Temporary pass 1!", "form_token": signupToken(t, app),
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	// Simulate an imported legacy account: Better Auth hash, verified email.
	if _, err := db.ExecContext(
		ctx,
		`UPDATE users SET password = ?, email_verified_at = CURRENT_TIMESTAMP WHERE email = ?`,
		f.Hash,
		email,
	); err != nil {
		t.Fatal(err)
	}

	c.PostJSON("/login", map[string]any{"identifier": email, "password": f.Password + "x"}).AssertRedirect("/login")

	var stored string
	read := func() string {
		if err := db.QueryRowContext(ctx, `SELECT password FROM users WHERE email = ?`, email).
			Scan(&stored); err != nil {
			t.Fatal(err)
		}
		return stored
	}
	if read() != f.Hash {
		t.Fatal("failed login must not touch the stored hash")
	}

	c.PostJSON("/login", map[string]any{"identifier": email, "password": f.Password}).AssertRedirect("/onboarding")
	if got := read(); !strings.HasPrefix(got, "$argon2id$") {
		t.Fatalf("stored hash after login = %.12q..., want $argon2id$ prefix", got)
	}
}

// signupToken is a sign-up form token rendered forms.MinFill ago.
func signupToken(t *testing.T, app *vukemtest.App) string {
	t.Helper()

	key, err := app.DeriveKey("signup.form", 32)
	if err != nil {
		t.Fatal(err)
	}

	return forms.Token{Key: key, Now: app.Now}.IssueAt("signup", app.Now().Add(-forms.MinFill))
}
