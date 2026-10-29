package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/AlexGalhardo/vukem/uuid"
)

// Retention is how many dumps of each kind stay in the bucket (SDD 11 §7.3
// step 5). Buckets have no lifecycle rules, so pruning is app code.
var Retention = map[string]int{"daily": 30, "monthly": 12}

// PruneSelection returns the keys to delete so that the newest keep remain.
// Keys sort by time: the date path and the UTC stamp come before anything else
// that varies within one kind and environment.
func PruneSelection(keys []string, keep int) []string {
	s := slices.Clone(keys)
	slices.Sort(s)

	if len(s) <= keep {
		return nil
	}

	return s[:len(s)-keep]
}

// Prune deletes the dumps beyond the retention of each kind and returns how many.
func (r *Runner) Prune(ctx context.Context) (int, error) {
	deleted := 0

	for _, kind := range Kinds {
		keys, err := r.Disk.List(ctx, "backups/postgres/"+kind)
		if err != nil {
			return deleted, fmt.Errorf("backup: list %s: %w", kind, err)
		}

		for _, k := range PruneSelection(keys, Retention[kind]) {
			if err := r.Disk.Delete(ctx, k); err != nil {
				return deleted, fmt.Errorf("backup: prune %s: %w", k, err)
			}

			deleted++
		}
	}

	return deleted, nil
}

// CopyMonthly copies a verified daily dump to the monthly prefix. The frames
// are bound to their object key, so the copy is decrypted and sealed again for
// the new key, in a stream; the plaintext never touches disk.
func (r *Runner) CopyMonthly(ctx context.Context, src Result) (Result, error) {
	started := r.Now().UTC()
	res := Result{ID: uuid.NewV7().String(), Key: ObjectKey("monthly", r.Env, started)}

	if _, err := r.DB.ExecContext(ctx, `INSERT INTO backup_runs (id, kind, object_key, status, started_at)
VALUES ($1, 'monthly', $2, 'running', $3)`, res.ID, res.Key, started); err != nil {
		return res, fmt.Errorf("backup: record monthly copy: %w", err)
	}

	obj, err := r.Disk.Open(ctx, src.Key)
	if err != nil {
		return res, errors.Join(err, r.finish(context.WithoutCancel(ctx), res, started, err))
	}
	defer func() { _ = obj.Close() }()

	sum := sha256.New()
	size := &counter{}
	pr, pw := io.Pipe()
	copied := make(chan error, 1)

	go func() {
		fw := NewFrameWriter(io.MultiWriter(pw, sum, size), r.Enc, res.Key)
		_, err := io.Copy(fw, NewFrameReader(obj, r.Enc, src.Key))
		err = errors.Join(err, fw.Close())
		_ = pw.CloseWithError(err)
		copied <- err
	}()

	putErr := r.Disk.Put(ctx, res.Key, pr)
	_ = pr.CloseWithError(errors.New("backup: upload stopped"))

	runErr := errors.Join(<-copied, putErr)
	res.SHA256, res.Bytes = hex.EncodeToString(sum.Sum(nil)), size.n

	if runErr != nil {
		_ = r.Disk.Delete(context.WithoutCancel(ctx), res.Key)
	}

	return res, errors.Join(runErr, r.finish(context.WithoutCancel(ctx), res, started, runErr))
}
