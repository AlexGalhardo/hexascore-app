package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Retention is one operational table the daily prune trims (SDD 11 §5, audit T-17): rows whose Column is older
// than Days are deleted. Where narrows the rows that may go (ingest_failures keeps open failures).
type Retention struct {
	Table, Column, Where string
	Days                 int
}

// Retentions are the windows of the operational tables that had none (audit T-17). Each can be overridden with
// RETENTION_<TABLE>_DAYS (a positive integer).
var Retentions = []Retention{
	{Table: "contact_messages", Column: "created_at", Days: 180},
	{Table: "suggestion_events", Column: "created_at", Days: 365},
	{Table: "ingest_failures", Column: "last_seen", Where: "parked", Days: 90},
	{Table: "backup_runs", Column: "started_at", Days: 365},
	{Table: "admin_audit_logs", Column: "created_at", Days: 730},
}

// RetentionWindows applies the RETENTION_<TABLE>_DAYS overrides to Retentions; a value that is not a positive
// integer is an error, so a typo fails at boot instead of keeping rows forever or deleting too much.
func RetentionWindows(get func(string) string) ([]Retention, error) {
	out := make([]Retention, len(Retentions))
	copy(out, Retentions)

	for i, r := range out {
		key := "RETENTION_" + strings.ToUpper(r.Table) + "_DAYS"
		raw := strings.TrimSpace(get(key))
		if raw == "" {
			continue
		}

		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("%s must be a positive integer, got %q", key, raw)
		}

		out[i].Days = n
	}

	return out, nil
}
