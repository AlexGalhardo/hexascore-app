package models

import (
	"database/sql"
	"time"

	"github.com/AlexGalhardo/vukem/orm"
)

// JobRun is one ingestion run (SDD 06 §5). Read-only in the admin panel.
type JobRun struct {
	ID          int64
	Job         string
	Source      sql.NullString
	Mode        string
	Status      string
	Seen        int
	Changed     int
	Failed      int
	Skipped     int
	RateLimited int
	TriggeredBy string
	Error       sql.NullString
	StartedAt   orm.NullTime
	FinishedAt  orm.NullTime
	CreatedAt   time.Time
}

// IngestFailure is a dead-letter target. The real key is (source, external_id); the panel only lists it, so
// PrimaryKey names last_run_id to satisfy the ORM's single-column key: never call orm.Update or orm.Delete (VF-035).
type IngestFailure struct {
	Source     string
	ExternalID string
	LastRunID  int64
	ErrorClass string
	Error      string
	Attempts   int
	Parked     bool
	FirstSeen  time.Time
	LastSeen   time.Time
}

// PrimaryKey names last_run_id only to satisfy the ORM's single-column key.
func (IngestFailure) PrimaryKey() string { return "last_run_id" }

// DQResult is one assert result of a dq.asserts run. The real key is (run_id, code); list-only like IngestFailure.
type DQResult struct {
	RunID     int64
	Code      string
	Severity  string
	Passed    bool
	Observed  sql.NullFloat64
	Threshold sql.NullFloat64
	CheckedAt time.Time
}

// Table names the dq_results table (the default would split the DQ prefix).
func (DQResult) Table() string { return "dq_results" }

// PrimaryKey names run_id only to satisfy the ORM's single-column key.
func (DQResult) PrimaryKey() string { return "run_id" }
