package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/AlexGalhardo/vukem/orm"
)

// Suggestion is one community edit, create or merge request (SDD 04 §5).
// The id is a UUID v7 set in Go. User ids are bigint, matching users.id.
type Suggestion struct {
	ID           string
	AuthorID     sql.NullInt64
	Kind         string
	TargetType   string
	TargetID     sql.NullString
	TargetDomain string
	Locale       string
	Changes      orm.JSON[json.RawMessage]
	Note         sql.NullString
	SourceURLs   TextArray `db:"source_urls"`
	Status       string
	AssignedTo   sql.NullInt64
	ClaimedAt    orm.NullTime
	DecidedBy    sql.NullInt64
	DecidedAt    orm.NullTime
	RejectReason sql.NullString
	DecisionNote sql.NullString
	SubmittedAt  orm.NullTime
	DependsOn    sql.NullString
	Meta         orm.JSON[json.RawMessage]
	Version      int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SuggestionEvent is one status transition of a Suggestion (SDD 04 §5).
type SuggestionEvent struct {
	ID           int64
	SuggestionID string
	FromStatus   sql.NullString
	ToStatus     string
	ActorID      sql.NullInt64
	ActorRole    string
	Reason       sql.NullString
	Meta         orm.JSON[json.RawMessage]
	CreatedAt    time.Time
}

// ContributorStats is the rebuildable per-user accuracy row (SDD 04 §5).
type ContributorStats struct {
	UserID          int64
	SubmittedTotal  int
	ApprovedTotal   int
	RejectedTotal   int
	RevertedTotal   int
	WithdrawnTotal  int
	Tier            string
	CooldownUntil   orm.NullTime
	BlockedUntil    orm.NullTime
	TrustedEligible bool
	UpdatedAt       time.Time
}
