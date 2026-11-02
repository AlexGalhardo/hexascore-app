package store

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// HistoryTable is the partitioned table of superseded source payloads (ADR 0025).
const HistoryTable = "source_record_history"

var partitionName = regexp.MustCompile(`^source_record_history_(\d{4})_(\d{2})$`)

// HistoryPartition is one monthly partition of HistoryTable.
type HistoryPartition struct {
	Name string
	// From (inclusive) and To (exclusive) are the UTC month bounds of superseded_at.
	From, To time.Time
}

// Month is the "YYYY-MM" label used in export object keys.
func (p HistoryPartition) Month() string { return p.From.Format("2006-01") }

func monthStart(t time.Time) time.Time {
	t = t.UTC()

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// EnsureHistoryPartitions creates the monthly partitions from the month before now through ahead months after it
// (names are built from integers, never from input). A month already in the default partition cannot get its own;
// that error is returned, which is why the job runs this daily and the migration pre-creates three months.
func EnsureHistoryPartitions(ctx context.Context, ex database.Executor, now time.Time, ahead int) error {
	for m := monthStart(now).AddDate(0, -1, 0); !m.After(monthStart(now).AddDate(0, ahead, 0)); m = m.AddDate(0, 1, 0) {
		q := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s_%s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			HistoryTable, m.Format("2006_01"), HistoryTable, m.Format("2006-01-02 15:04:05+00"),
			m.AddDate(0, 1, 0).Format("2006-01-02 15:04:05+00"))
		if _, err := ex.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("store: partition %s: %w", m.Format("2006-01"), err)
		}
	}

	return nil
}

// HistoryPartitions lists the monthly partitions, oldest first (the default partition is not one).
func HistoryPartitions(ctx context.Context, ex database.Executor) ([]HistoryPartition, error) {
	rows, err := ex.QueryContext(ctx, `SELECT c.relname FROM pg_inherits i
JOIN pg_class c ON c.oid = i.inhrelid JOIN pg_class p ON p.oid = i.inhparent
WHERE p.relname = $1 ORDER BY c.relname`, HistoryTable)
	if err != nil {
		return nil, fmt.Errorf("store: list history partitions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []HistoryPartition

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		m := partitionName.FindStringSubmatch(name)
		if m == nil {
			continue
		}

		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		from := time.Date(y, time.Month(mo), 1, 0, 0, 0, 0, time.UTC)

		out = append(out, HistoryPartition{Name: name, From: from, To: from.AddDate(0, 1, 0)})
	}

	return out, rows.Err()
}

// DropHistoryPartition drops a monthly partition returned by HistoryPartitions.
func DropHistoryPartition(ctx context.Context, ex database.Executor, p HistoryPartition) error {
	if !partitionName.MatchString(p.Name) {
		return fmt.Errorf("store: %q is not a history partition", p.Name)
	}

	if _, err := ex.ExecContext(ctx, `DROP TABLE `+p.Name); err != nil {
		return fmt.Errorf("store: drop %s: %w", p.Name, err)
	}

	return nil
}
