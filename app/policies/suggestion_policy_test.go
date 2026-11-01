package policies_test

import (
	"database/sql"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/policies"
)

func TestStatusPagePolicy(t *testing.T) {
	t.Parallel()

	s := &models.Suggestion{AuthorID: sql.NullInt64{Int64: 1, Valid: true}}
	orphan := &models.Suggestion{}
	author := &models.User{ID: 1}
	other := &models.User{ID: 2}
	admin := &models.User{ID: 3, Role: models.RoleAdmin}
	p := policies.SuggestionPolicy{}

	cases := []struct {
		name         string
		u            *models.User
		s            *models.Suggestion
		view, author bool
	}{
		{"author", author, s, true, true},
		{"admin", admin, s, true, false},
		{"other", other, s, false, false},
		{"anonymous", nil, s, false, false},
		{"deleted author, admin", admin, orphan, true, false},
		{"deleted author, user", other, orphan, false, false},
	}
	for _, c := range cases {
		if got := p.View(t.Context(), c.u, c.s).Allowed(); got != c.view {
			t.Errorf("%s: view = %v, want %v", c.name, got, c.view)
		}
		if got := p.Submit(t.Context(), c.u, c.s).Allowed(); got != c.author {
			t.Errorf("%s: submit = %v, want %v", c.name, got, c.author)
		}
		if got := p.Withdraw(t.Context(), c.u, c.s).Allowed(); got != c.author {
			t.Errorf("%s: withdraw = %v, want %v", c.name, got, c.author)
		}
	}
}
