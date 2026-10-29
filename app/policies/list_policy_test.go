package policies_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/authz"
	"github.com/AlexGalhardo/vukem/errs"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/policies"
)

func user(id int64) *models.User {
	u := &models.User{}
	u.ID = id

	return u
}

// TestSecurity_A01_ListViewPolicy: a private, hidden or owner-private list is a 404 for everyone but its owner
// (SDD 14 §6.1); public visible lists are readable anonymously.
func TestSecurity_A01_ListViewPolicy(t *testing.T) {
	base := models.List{ID: "l", UserID: 1, IsPublic: true}
	cases := []struct {
		name   string
		list   models.List
		viewer *models.User
		want   int
	}{
		{"anonymous public", base, nil, 0},
		{"stranger public", base, user(2), 0},
		{"stranger private", models.List{UserID: 1}, user(2), http.StatusNotFound},
		{"anonymous private", models.List{UserID: 1}, nil, http.StatusNotFound},
		{"stranger hidden", models.List{UserID: 1, IsPublic: true, Hidden: true}, user(2), http.StatusNotFound},
		{
			"stranger owner-private",
			models.List{UserID: 1, IsPublic: true, OwnerPrivate: true},
			user(2),
			http.StatusNotFound,
		},
		{"owner private", models.List{UserID: 1, Hidden: true, OwnerPrivate: true}, user(1), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gate := authz.New(func(context.Context) (*models.User, error) { return tc.viewer, nil })
			if err := gate.Register(authz.For[*models.List](policies.ListPolicy{})); err != nil {
				t.Fatal(err)
			}

			err := gate.Authorize(t.Context(), "view", &tc.list)
			if got := errs.Status(err); (err == nil) != (tc.want == 0) || (err != nil && got != tc.want) {
				t.Fatalf("view: err %v status %d, want %d", err, got, tc.want)
			}

			owner := tc.viewer != nil && tc.viewer.ID == tc.list.UserID
			for _, ability := range []string{"update", "delete", "addItem", "reorder"} {
				err := gate.Authorize(t.Context(), ability, &tc.list)
				if owner != (err == nil) || (!owner && errs.Status(err) != http.StatusNotFound) {
					t.Errorf("%s: err %v, owner %t", ability, err, owner)
				}
			}
		})
	}
}
