package tests

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestSecurity_A01_AbilityRequired: a token missing the route ability is 403
// on the shipped /app/v1 write (SDD 05 §6.4). No token is 401.
func TestSecurity_A01_AbilityRequired(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	u := makeUser(t, app, "ability@example.com", "user", true)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	plain, _, err := acc.Tokens.Create(t.Context(), u.AuthID(), "ro", []string{"status:write"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	const path = "/app/v1/items/0197f3a0-0000-7000-8000-000000000000/rating"
	body := map[string]any{"scores": map[string]float64{"gameplay": 7.5}}
	app.Client(t).WithToken(plain).PutJSON(path, body).AssertForbidden()
	app.Client(t).PutJSON(path, body).AssertUnauthorized()
}
