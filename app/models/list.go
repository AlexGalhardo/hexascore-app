package models

import "time"

// List is a user's ordered list of items of one type (SDD 14 §5). It is read with raw SQL, not the ORM, so
// OwnerPrivate can ride along for the policy.
type List struct {
	ID          string
	UserID      int64
	Type        string
	Title       string
	Description *string
	IsPublic    bool
	// Hidden is the moderation flag; only an admin sets it.
	Hidden    bool
	ItemCount int
	// Version is bumped by every item add and reorder.
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	// OwnerPrivate is users.profile_private of the owner; it hides public lists too.
	OwnerPrivate bool
}
