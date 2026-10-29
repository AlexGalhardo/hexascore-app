package models

import (
	"database/sql"
	"time"

	"github.com/AlexGalhardo/vukem/orm"
)

// Report is a member report on a list or a user (SDD 14 §5). IDs are uuid v7 strings set in Go; the admin panel
// reads it and outcomes go through POST /admin/reports/{id}/resolve.
type Report struct {
	ID         string
	ReporterID sql.NullInt64
	TargetType string
	TargetID   string
	Reason     string
	Details    *string
	Status     string
	Outcome    sql.NullString
	AdminNote  *string
	ResolvedBy sql.NullInt64
	ResolvedAt orm.NullTime
	CreatedAt  time.Time
}

// ContactMessage is one message of the contact form (SDD 14 §5); only Handled is writable in the panel.
type ContactMessage struct {
	ID        string
	UserID    sql.NullInt64
	Name      string
	Email     string
	Subject   string
	Message   string
	Handled   bool
	CreatedAt time.Time
}
