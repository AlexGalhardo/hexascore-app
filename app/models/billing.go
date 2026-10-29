package models

import (
	"database/sql"

	"github.com/AlexGalhardo/vukem/orm"
)

// Billable is the Stripe side of a user, kept in the users table (vukem
// install:billing). It is a separate model so User stays as it is.
type Billable struct {
	orm.Model

	StripeCustomerID sql.Null[string] `json:"-"`
}

// Table maps Billable to the users table.
func (Billable) Table() string { return "users" }

// Subscription mirrors a Stripe subscription; webhooks keep it in sync.
type Subscription struct {
	orm.Model

	UserID               int64        `json:"-"`
	StripeSubscriptionID string       `json:"-"`
	Status               string       `json:"status"`
	PriceID              string       `json:"-"`
	Plan                 string       `json:"plan"`
	Provider             string       `json:"-"`
	State                string       `json:"state"`
	CurrentPeriodEnd     orm.NullTime `json:"current_period_end"`
	CancelAtPeriodEnd    bool         `json:"cancel_at_period_end"`
}

// Valid reports whether the subscription grants access: its state is active
// (Stripe active or trialing). past_due, unpaid, incomplete, paused and
// canceled do not (SDD 05 section 7.5).
func (s Subscription) Valid() bool { return s.State == "active" }
