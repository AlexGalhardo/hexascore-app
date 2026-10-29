package backup

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/logging"
)

// DrillDatabase is the scratch database the drill restores into and drops.
const DrillDatabase = "hexascore_drill"

// Restore streams the dump at key into the database at targetURL: open,
// decrypt and pipe into pg_restore, so the plaintext never touches disk. The
// archive arrives on stdin, which rules out pg_restore -j; one job restores
// the 3 GB estimate well inside the 60 min drill budget (SDD 11 §7.4).
func (r *Runner) Restore(ctx context.Context, key, targetURL string) error {
	env, err := pgEnv(targetURL)
	if err != nil {
		return err
	}

	obj, err := r.Disk.Open(ctx, key)
	if err != nil {
		return fmt.Errorf("backup: open dump: %w", err)
	}
	defer func() { _ = obj.Close() }()

	stderr := &capped{max: 4 << 10}
	// Without -d pg_restore writes SQL to stdout instead of connecting; the
	// name repeats PGDATABASE, the rest of the connection comes from env.
	db := ""
	if u, err := url.Parse(targetURL); err == nil {
		db = strings.TrimPrefix(u.Path, "/")
	}

	//nolint:gosec // G204: binary and args set in code
	cmd := exec.CommandContext(ctx, r.PgRestore, "--no-owner", "--no-acl", "--exit-on-error", "-d", db)
	cmd.Env, cmd.Stdin, cmd.Stderr = env, NewFrameReader(obj, r.Enc, key), stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup: pg_restore: %w: %s", err, bytes.TrimSpace(stderr.buf.Bytes()))
	}

	return nil
}

// LatestDump returns the newest dump of kind in the disk.
func (r *Runner) LatestDump(ctx context.Context, kind string) (string, error) {
	keys, err := r.Disk.List(ctx, "backups/postgres/"+kind)
	if err != nil {
		return "", fmt.Errorf("backup: list %s: %w", kind, err)
	}

	keys = slices.DeleteFunc(keys, func(k string) bool { return !strings.HasSuffix(k, ".dump.enc") })
	if len(keys) == 0 {
		return "", fmt.Errorf("backup: no %s dump in the bucket", kind)
	}

	slices.Sort(keys)

	return keys[len(keys)-1], nil
}

// DrillResult is one restore drill.
type DrillResult struct {
	ID, Key  string
	Duration time.Duration
	// Diff lists tables whose restored count differs from the manifest.
	Diff []string
	// AssertErrors counts error-level data quality violations.
	AssertErrors int
}

// DrillBudget is the FR-11-6 limit for one drill.
const DrillBudget = time.Hour

// Drill restores the latest daily dump into a fresh DrillDatabase on the
// server of adminURL, compares row counts with the dump manifest, runs the
// data quality asserts and drops the database (SDD 11 §7.4). It never
// touches the database adminURL names. The run is recorded in r.DB as
// backup_runs(kind='drill'): verified when restore, counts and asserts pass
// within DrillBudget, failed otherwise.
func (r *Runner) Drill(
	ctx context.Context,
	adminURL string,
	asserts func(ctx context.Context, db *database.DB) (int, error),
) (DrillResult, error) {
	started := r.Now().UTC()
	res := DrillResult{ID: uuid.NewV7().String()}

	if _, err := r.DB.ExecContext(ctx, `INSERT INTO backup_runs (id, kind, status, started_at)
VALUES ($1, 'drill', 'running', $2)`, res.ID, started); err != nil {
		return res, fmt.Errorf("backup: record drill start: %w", err)
	}

	err := r.drill(ctx, adminURL, asserts, &res)
	finished := r.Now().UTC()
	res.Duration = finished.Sub(started)

	if err == nil && res.Duration > DrillBudget {
		err = fmt.Errorf("backup: drill took %s, budget %s", res.Duration.Round(time.Second), DrillBudget)
	}

	status, msg := "verified", sql.NullString{}
	if err != nil {
		s := logging.Scrub(err.Error())
		status, msg = "failed", sql.NullString{String: s[:min(len(s), 2000)], Valid: true}
	}

	if _, dbErr := r.DB.ExecContext(context.WithoutCancel(ctx), `UPDATE backup_runs SET object_key = $2, status = $3,
error = $4, duration_s = $5, finished_at = $6 WHERE id = $1`,
		res.ID, sql.NullString{String: res.Key, Valid: res.Key != ""}, status, msg,
		int(res.Duration.Seconds()), finished); dbErr != nil {
		return res, errors.Join(err, fmt.Errorf("backup: record drill: %w", dbErr))
	}

	return res, err
}

func (r *Runner) drill(
	ctx context.Context,
	adminURL string,
	asserts func(ctx context.Context, db *database.DB) (int, error),
	res *DrillResult,
) (err error) {
	if res.Key, err = r.LatestDump(ctx, "daily"); err != nil {
		return err
	}

	manifest, err := r.ReadManifest(ctx, res.Key)
	if err != nil {
		return err
	}

	targetURL, err := withDatabase(adminURL, DrillDatabase)
	if err != nil {
		return err
	}

	admin, err := database.Open(ctx, adminURL, database.WithDriver(postgres.Driver()))
	if err != nil {
		return fmt.Errorf("backup: connect: %w", err)
	}
	defer func() { _ = admin.Close() }()

	// A drill killed half way leaves the database behind; start clean and always drop it.
	drop := `DROP DATABASE IF EXISTS ` + DrillDatabase + ` WITH (FORCE)`
	if _, err := admin.ExecContext(ctx, drop); err != nil {
		return fmt.Errorf("backup: drop %s: %w", DrillDatabase, err)
	}

	if _, err := admin.ExecContext(ctx, `CREATE DATABASE `+DrillDatabase); err != nil {
		return fmt.Errorf("backup: create %s: %w", DrillDatabase, err)
	}

	defer func() {
		if _, dropErr := admin.ExecContext(context.WithoutCancel(ctx), drop); dropErr != nil {
			err = errors.Join(err, fmt.Errorf("backup: drop %s: %w", DrillDatabase, dropErr))
		}
	}()

	if err := r.Restore(ctx, res.Key, targetURL); err != nil {
		return err
	}

	target, err := database.Open(ctx, targetURL, database.WithDriver(postgres.Driver()))
	if err != nil {
		return fmt.Errorf("backup: connect %s: %w", DrillDatabase, err)
	}
	defer func() { _ = target.Close() }()

	counts, err := CountRows(ctx, target)
	if err != nil {
		return err
	}

	if res.Diff = CountDiff(manifest.Tables, counts); len(res.Diff) > 0 {
		return fmt.Errorf("backup: restored counts differ from the manifest: %s", strings.Join(res.Diff, "; "))
	}

	if res.AssertErrors, err = asserts(ctx, target); err != nil {
		return fmt.Errorf("backup: data quality asserts: %w", err)
	}

	if res.AssertErrors > 0 {
		return fmt.Errorf("backup: %d data quality assert errors", res.AssertErrors)
	}

	return nil
}

// withDatabase returns dsn pointing at database name on the same server.
func withDatabase(dsn, name string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil || u.Host == "" {
		return "", errors.New("backup: database URL is not a postgres:// URL")
	}

	u.Path = "/" + name

	return u.String(), nil
}
