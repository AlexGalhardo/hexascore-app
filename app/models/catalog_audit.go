package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/AlexGalhardo/vukem/orm"
)

// CatalogAudit is one append-only catalogue change (SDD 04 §5). The database
// rejects UPDATE and DELETE. User ids are bigint, matching users.id.
type CatalogAudit struct {
	ID             int64
	OccurredAt     time.Time
	ActorID        sql.NullInt64
	ActorRole      string
	Source         string
	SuggestionID   sql.NullString
	TargetType     string
	TargetID       string
	Field          string
	Before         orm.JSON[json.RawMessage]
	After          orm.JSON[json.RawMessage]
	PrevProvenance orm.JSON[json.RawMessage]
	Reason         sql.NullString
	RevertsAuditID sql.NullInt64
	RequestID      sql.NullString
}

// CatalogLock blocks drafts, submits and approvals on one field or on "*".
type CatalogLock struct {
	ID         string
	TargetType string
	TargetID   string
	Field      string
	Reason     string
	LockedBy   sql.NullInt64
	ExpiresAt  orm.NullTime
	CreatedAt  time.Time
}
