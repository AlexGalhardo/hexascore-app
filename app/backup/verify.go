package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/logging"
)

// ErrSmallDump fails a dump whose size is at most half the median of the
// verified dumps of the same kind in the previous 7 days (SDD 11 §7.3 step 4).
var ErrSmallDump = errors.New("backup: dump is at most 50% of the 7-day median")

// requiredTables must appear in the dump's table of contents.
var requiredTables = []string{"items", "users", "ratings"}

// CheckSize applies the size rule; with no earlier dumps there is nothing to compare.
func CheckSize(size int64, recent []int64) error {
	if len(recent) == 0 {
		return nil
	}

	s := slices.Clone(recent)
	slices.Sort(s)

	median := float64(s[len(s)/2])
	if len(s)%2 == 0 {
		median = float64(s[len(s)/2-1]+s[len(s)/2]) / 2
	}

	if float64(size) <= median/2 {
		return fmt.Errorf("%w: %d bytes, median %.0f", ErrSmallDump, size, median)
	}

	return nil
}

// MissingTables returns the required tables absent from a pg_restore --list output.
func MissingTables(toc []byte) []string {
	var missing []string

	for _, t := range requiredTables {
		// TOC lines look like "123; 1259 16390 TABLE public items owner".
		if !regexp.MustCompile(`(?m)^\d+; \d+ \d+ TABLE public ` + t + ` `).Match(toc) {
			missing = append(missing, t)
		}
	}

	return missing
}

// Verify checks an uploaded dump: size against the 7-day median, then decrypt
// and pg_restore --list must name items, users and ratings. The row becomes
// verified, or failed with the scrubbed reason.
func (r *Runner) Verify(ctx context.Context, kind string, res Result) error {
	err := r.verify(ctx, kind, res)

	status, msg := "verified", sql.NullString{}
	if err != nil {
		s := logging.Scrub(err.Error())
		status, msg = "failed", sql.NullString{String: s[:min(len(s), 2000)], Valid: true}
	}

	if _, dbErr := r.DB.ExecContext(context.WithoutCancel(ctx), `UPDATE backup_runs SET status = $2, error = $3
WHERE id = $1`, res.ID, status, msg); dbErr != nil {
		return errors.Join(err, fmt.Errorf("backup: record verify: %w", dbErr))
	}

	return err
}

func (r *Runner) verify(ctx context.Context, kind string, res Result) error {
	since := r.Now().UTC().Add(-7 * 24 * time.Hour)

	rows, err := r.DB.QueryContext(ctx, `SELECT bytes FROM backup_runs WHERE kind = $1 AND status = 'verified'
AND started_at >= $2 AND id <> $3 AND bytes IS NOT NULL`, kind, since, res.ID)
	if err != nil {
		return fmt.Errorf("backup: recent sizes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var recent []int64

	for rows.Next() {
		var b int64
		if err := rows.Scan(&b); err != nil {
			return fmt.Errorf("backup: recent sizes: %w", err)
		}

		recent = append(recent, b)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("backup: recent sizes: %w", err)
	}

	if err := CheckSize(res.Bytes, recent); err != nil {
		return err
	}

	obj, err := r.Disk.Open(ctx, res.Key)
	if err != nil {
		return fmt.Errorf("backup: open dump: %w", err)
	}
	defer func() { _ = obj.Close() }()

	stderr := &capped{max: 4 << 10}
	// pg_restore --list reads the archive from stdin and needs no connection, so it gets PATH only.
	//nolint:gosec // G204: binary and args set in code
	cmd := exec.CommandContext(ctx, r.PgRestore, "--list")
	cmd.Env, cmd.Stdin, cmd.Stderr = []string{binPath()}, NewFrameReader(obj, r.Enc, res.Key), stderr

	toc, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("backup: pg_restore --list: %w: %s", err, stderr.buf.Bytes())
	}

	if missing := MissingTables(toc); len(missing) > 0 {
		return fmt.Errorf("backup: dump lacks tables %v", missing)
	}

	return nil
}
