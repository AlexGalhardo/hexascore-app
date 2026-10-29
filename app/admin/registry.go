package admin

import (
	"time"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/authz"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Resources lists the panel's resources; vukem make:admin-resource appends
// here.
func Resources() []vadmin.AnyResource {
	return []vadmin.AnyResource{UsersResource()}
}

// Policies lists what admins may do per model; vukem make:admin-resource
// appends here. A resource without a policy denies every action.
func Policies(now func() time.Time) []authz.Policy {
	return []authz.Policy{authz.For[*models.User](UserPolicy{now: now})}
}
