// Package privacy holds LGPD retention (SDD 13 §5, §7.6).
package privacy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

const (
	// SessionClientMaxAge is how long a session may keep the recorded IP and user agent.
	SessionClientMaxAge = 30 * 24 * time.Hour
	// ClientReportMaxAge is how long a client_reports row may live (SDD 12).
	// pruneClientReports no-ops until the table exists.
	ClientReportMaxAge = 7 * 24 * time.Hour
)

// Expirer deletes export ZIPs past their 24 h link. *library.Exports implements it.
type Expirer interface {
	Expire(ctx context.Context) (int, error)
}

// Result counts what one retention run changed.
type Result struct {
	SessionsCleared int
	ReportsDeleted  int64
	ExportsExpired  int
}

// Run strips session IP and user agent older than 30 days, deletes client_reports older than 7 days when
// that table exists, and expires export ZIPs. now is the app clock. Non-Postgres is a no-op.
func Run(ctx context.Context, db *database.DB, expire Expirer, now time.Time) (Result, error) {
	var out Result
	if db == nil || db.Dialect().Name() != "postgres" {
		return out, nil
	}

	now = now.UTC()

	n, err := clearSessionClients(ctx, db, now)
	out.SessionsCleared = n
	if err != nil {
		return out, err
	}

	deleted, err := pruneClientReports(ctx, db, now)
	out.ReportsDeleted = deleted
	if err != nil {
		return out, err
	}

	if expire != nil {
		expired, err := expire.Expire(ctx)
		out.ExportsExpired = expired
		if err != nil {
			return out, fmt.Errorf("privacy: expire exports: %w", err)
		}
	}

	return out, nil
}

func tableExists(ctx context.Context, db *database.DB, name string) (bool, error) {
	var got sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT to_regclass($1)::text`, name).Scan(&got); err != nil {
		return false, fmt.Errorf("privacy: lookup %s: %w", name, err)
	}

	return got.Valid && got.String != "", nil
}

func clearSessionClients(ctx context.Context, db *database.DB, now time.Time) (int, error) {
	ok, err := tableExists(ctx, db, "sessions")
	if err != nil || !ok {
		return 0, err
	}

	// strpos, not LIKE: "_" is a LIKE wildcard, and the key is "_client.".
	rows, err := db.QueryContext(ctx, `SELECT id, data FROM sessions WHERE strpos(data, '"_client.') > 0`)
	if err != nil {
		return 0, fmt.Errorf("privacy: list sessions: %w", err)
	}
	defer rows.Close()

	cutoff := now.Add(-SessionClientMaxAge).Unix()
	cleared := 0

	for rows.Next() {
		var id, raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return cleared, fmt.Errorf("privacy: scan session: %w", err)
		}

		next, changed, err := stripClient(raw, cutoff)
		if err != nil || !changed {
			if err != nil {
				return cleared, fmt.Errorf("privacy: session %s: %w", id, err)
			}

			continue
		}

		if _, err := db.ExecContext(ctx, `UPDATE sessions SET data = $1 WHERE id = $2`, next, id); err != nil {
			return cleared, fmt.Errorf("privacy: update session %s: %w", id, err)
		}

		cleared++
	}

	if err := rows.Err(); err != nil {
		return cleared, fmt.Errorf("privacy: sessions: %w", err)
	}

	return cleared, nil
}

// stripClient removes _client.ip and _client.ua when the session started at or before cutoff (unix seconds).
// A missing or unreadable start time is treated as expired: an IP we cannot date is not kept.
// Other keys are left as raw JSON so integers stay exact.
func stripClient(raw string, cutoff int64) (string, bool, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return "", false, err
	}

	if _, ok := data["_client.ip"]; !ok {
		if _, ok := data["_client.ua"]; !ok {
			return "", false, nil
		}
	}

	if young, ok := sessionYoung(data["_created"], cutoff); ok && young {
		return "", false, nil
	}

	delete(data, "_client.ip")
	delete(data, "_client.ua")

	next, err := json.Marshal(data)
	if err != nil {
		return "", false, err
	}

	return string(next), true, nil
}

func sessionYoung(raw json.RawMessage, cutoff int64) (bool, bool) {
	if len(raw) == 0 {
		return false, false
	}

	var created int64
	if err := json.Unmarshal(raw, &created); err != nil {
		return false, false
	}

	return created > cutoff, true
}

func pruneClientReports(ctx context.Context, db *database.DB, now time.Time) (int64, error) {
	ok, err := tableExists(ctx, db, "client_reports")
	if err != nil || !ok {
		return 0, err
	}

	res, err := db.ExecContext(ctx,
		`DELETE FROM client_reports WHERE received_at < $1`, now.Add(-ClientReportMaxAge))
	if err != nil {
		return 0, fmt.Errorf("privacy: prune client_reports: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("privacy: prune client_reports: %w", err)
	}

	return n, nil
}
