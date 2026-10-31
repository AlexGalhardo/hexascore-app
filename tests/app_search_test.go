//go:build integration

package tests

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
)

// TestAppSearch covers optional-auth site search on /app/v1.
func TestAppSearch(t *testing.T) {
	app, _ := searchApp(t)
	anon := app.Client(t)

	anon.GetJSON("/app/v1/search?q=zelda&type=game").AssertOK().
		AssertJSONPath("data.0.title", "Zelda").
		AssertJSONCount("data", 3)
	anon.GetJSON("/app/v1/search?q=zel").AssertOK().AssertJSONPath("data.0.title", "Zelda")
	anon.GetJSON("/app/v1/search/suggest?q=zel").AssertOK().AssertJSONPath("data.0.title", "Zelda")

	long := strings.Repeat("a", 101)
	anon.GetJSON("/app/v1/search?q=" + long).AssertStatus(http.StatusBadRequest)
	anon.GetJSON("/app/v1/search/suggest?q=" + long).AssertStatus(http.StatusBadRequest)

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	u := makeUser(t, app, "search-app@example.com", "user", true)
	plain, _, err := acc.Tokens.Create(t.Context(), u.AuthID(), "phone", []string{"catalog:read"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	app.Client(t).WithToken(plain).GetJSON("/app/v1/search?q=zelda&type=game").AssertOK().
		AssertJSONPath("data.0.title", "Zelda")

	anon.PostJSON("/app/v1/search/events", map[string]any{
		"searchId": uuid.NewV7().String(),
		"refId":    uuid.NewV7().String(),
		"position": 1,
	}).AssertNoContent()
	anon.PostJSON("/app/v1/search/events", map[string]any{"searchId": "x"}).AssertUnprocessable()
}
