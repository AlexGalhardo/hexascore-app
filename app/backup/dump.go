package backup

import (
	"bytes"
	"cmp"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/AlexGalhardo/vukem/crypt"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/logging"
)

// Kinds backup:run writes; drill and offsite rows come from other jobs.
var Kinds = []string{"daily", "monthly"}

// Runner dumps, encrypts and uploads one backup (SDD 11 §7.3 steps 2-3).
type Runner struct {
	DB          *database.DB
	Disk        storage.Disk
	Enc         *crypt.Encrypter
	DatabaseURL string
	// PgDump is the pg_dump binary (the image ships postgresql18-client).
	PgDump string
	// PgRestore is the pg_restore binary Verify runs with --list.
	PgRestore string
	// Env names the environment in the object key.
	Env string
	Now func() time.Time
}

// Result is the recorded run.
type Result struct {
	ID, Key, SHA256 string
	Bytes           int64
}

// ObjectKey is backups/postgres/<kind>/YYYY/MM/DD/hexascore-<env>-<UTC stamp>.dump.enc.
func ObjectKey(kind, env string, at time.Time) string {
	at = at.UTC()

	return fmt.Sprintf("backups/postgres/%s/%s/hexascore-%s-%s.dump.enc",
		kind, at.Format("2006/01/02"), env, at.Format("20060102T1504Z"))
}

// Run streams pg_dump through the frame encrypter into the disk: the
// plaintext dump never touches disk. The row in backup_runs goes from
// running to uploaded or failed.
func (r *Runner) Run(ctx context.Context, kind string) (Result, error) {
	started := r.Now().UTC()
	res := Result{ID: uuid.NewV7().String(), Key: ObjectKey(kind, r.Env, started)}

	if _, err := r.DB.ExecContext(ctx, `INSERT INTO backup_runs (id, kind, object_key, status, started_at)
VALUES ($1, $2, $3, 'running', $4)`, res.ID, kind, res.Key, started); err != nil {
		return res, fmt.Errorf("backup: record start: %w", err)
	}

	snap, err := r.DB.SQL().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		err = fmt.Errorf("backup: snapshot: %w", err)

		return res, errors.Join(err, r.finish(context.WithoutCancel(ctx), res, started, err))
	}
	// The exported snapshot lives as long as this transaction: pg_dump must finish first.
	defer func() { _ = snap.Rollback() }()

	var snapshotID string

	if err := snap.QueryRowContext(ctx, `SELECT pg_export_snapshot()`).Scan(&snapshotID); err != nil {
		err = fmt.Errorf("backup: export snapshot: %w", err)

		return res, errors.Join(err, r.finish(context.WithoutCancel(ctx), res, started, err))
	}

	counts, err := CountRows(ctx, snap)
	if err != nil {
		return res, errors.Join(err, r.finish(context.WithoutCancel(ctx), res, started, err))
	}

	sum := sha256.New()
	size := &counter{}
	pr, pw := io.Pipe()
	dumped := make(chan error, 1)

	go func() {
		fw := NewFrameWriter(io.MultiWriter(pw, sum, size), r.Enc, res.Key)
		err := errors.Join(r.dump(ctx, fw, snapshotID), fw.Close())
		_ = pw.CloseWithError(err)
		dumped <- err
	}()

	putErr := r.Disk.Put(ctx, res.Key, pr)
	// Unblocks the dump when the upload stopped early.
	_ = pr.CloseWithError(errors.New("backup: upload stopped"))

	runErr := errors.Join(<-dumped, putErr)
	res.SHA256, res.Bytes = hex.EncodeToString(sum.Sum(nil)), size.n

	if runErr == nil {
		runErr = r.putManifest(ctx, Manifest{Key: res.Key, Tables: counts})
	}

	if runErr != nil {
		// A disk that kept a partial object (Local) must not leave it behind.
		_ = r.Disk.Delete(context.WithoutCancel(ctx), res.Key)
		_ = r.Disk.Delete(context.WithoutCancel(ctx), ManifestKey(res.Key))
	}

	return res, errors.Join(runErr, r.finish(context.WithoutCancel(ctx), res, started, runErr))
}

func (r *Runner) finish(ctx context.Context, res Result, started time.Time, runErr error) error {
	finished := r.Now().UTC()
	dur := int(finished.Sub(started).Seconds())

	var err error
	if runErr == nil {
		_, err = r.DB.ExecContext(ctx, `UPDATE backup_runs SET status = 'uploaded', bytes = $2, sha256 = $3,
duration_s = $4, finished_at = $5 WHERE id = $1`, res.ID, res.Bytes, res.SHA256, dur, finished)
	} else {
		msg := logging.Scrub(runErr.Error())
		_, err = r.DB.ExecContext(ctx, `UPDATE backup_runs SET status = 'failed', error = $2, duration_s = $3,
finished_at = $4 WHERE id = $1`, res.ID, msg[:min(len(msg), 2000)], dur, finished)
	}

	if err != nil {
		return fmt.Errorf("backup: record result: %w", err)
	}

	return nil
}

// dump runs pg_dump -Fc -Z 6 --no-owner into w, inside the exported snapshot
// the manifest was counted in. The child gets only the PG* variables it
// needs, never the app environment (secrets).
func (r *Runner) dump(ctx context.Context, w io.Writer, snapshot string) error {
	env, err := pgEnv(r.DatabaseURL)
	if err != nil {
		return err
	}

	stderr := &capped{max: 4 << 10}
	//nolint:gosec // G204: binary and args set in code
	cmd := exec.CommandContext(ctx, r.PgDump, "-Fc", "-Z", "6", "--no-owner", "--snapshot="+snapshot)
	cmd.Env, cmd.Stdout, cmd.Stderr = env, w, stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup: pg_dump: %w: %s", err, bytes.TrimSpace(stderr.buf.Bytes()))
	}

	return nil
}

// pgEnv maps a postgres:// URL to libpq variables.
func pgEnv(dsn string) ([]string, error) {
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
		return nil, errors.New("backup: DATABASE_URL is not a postgres:// URL")
	}

	pass, _ := u.User.Password()
	env := []string{
		"PGHOST=" + u.Hostname(),
		"PGPORT=" + cmp.Or(u.Port(), "5432"),
		"PGUSER=" + u.User.Username(),
		"PGPASSWORD=" + pass,
		"PGDATABASE=" + u.Path[min(1, len(u.Path)):],
		binPath(),
	}

	if mode := u.Query().Get("sslmode"); mode != "" {
		env = append(env, "PGSSLMODE="+mode)
	}

	return env, nil
}

// binPath passes PATH on: pg_restore resolves its own location from argv[0]
// through PATH ("could not find a pg_restore to execute" without it).
func binPath() string { return "PATH=" + os.Getenv("PATH") }

type counter struct{ n int64 }

func (c *counter) Write(p []byte) (int, error) { c.n += int64(len(p)); return len(p), nil }

// capped keeps the first max bytes of pg_dump's stderr for the error.
type capped struct {
	buf bytes.Buffer
	max int
}

func (c *capped) Write(p []byte) (int, error) {
	if room := c.max - c.buf.Len(); room > 0 {
		c.buf.Write(p[:min(room, len(p))])
	}

	return len(p), nil
}
