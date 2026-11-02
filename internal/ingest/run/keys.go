package run

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// Idempotency keys (RFC-001 section 4). workaround(vukem): VF-079, the queue has no unique-job option, so the app
// keeps ingest_job_keys. A key is sha256 of job, source, mode, params, cursor and the UTC day, so the daily chains
// are never blocked by the previous day's run, while a second automatic dispatch of the same stage on the same
// day is a no-op.
const (
	// KeyWindow is how long a finished dispatch blocks its twin.
	KeyWindow = "24 hours"
	// KeyStaleAfter frees a key whose job never finished (a crash lost it): after this long it may dispatch again.
	KeyStaleAfter = "3 hours"
	// KeyRetention is how long Prune keeps keys.
	KeyRetention = 7
)

// ErrDuplicate is returned by Create when the same automatic dispatch already happened; it wraps ErrRunActive so
// every caller that tolerates "already active" tolerates it.
var ErrDuplicate = fmt.Errorf("%w (same dispatch key)", ErrRunActive)

// keyed reports whether a run created by by goes through the key. Human and CLI starts and resumes always run.
func keyed(by string) bool {
	return strings.HasPrefix(by, "schedule") || strings.HasPrefix(by, "chain:")
}

// canonicalJSON re-encodes raw with sorted keys so equal documents hash alike; invalid or empty input is kept as is.
func canonicalJSON(raw json.RawMessage) string {
	var v any
	if len(raw) == 0 || json.Unmarshal(raw, &v) != nil {
		return string(raw)
	}

	b, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}

	return string(b)
}

// claimKey records the dispatch of run id and reports whether it may go: a new key, a failed one, a done one
// older than KeyWindow, or a queued/running one older than KeyStaleAfter. The day is the database's UTC day.
func claimKey(ctx context.Context, ex database.Executor, s Spec, id int64) (bool, error) {
	raw := strings.Join([]string{s.Job, s.Source, s.Mode, canonicalJSON(s.Params), canonicalJSON(s.Cursor)}, "|")

	var key string

	err := ex.QueryRowContext(ctx, `INSERT INTO ingest_job_keys (key, job_id, state)
VALUES (encode(sha256(convert_to($1 || '|' || (now() AT TIME ZONE 'UTC')::date::text, 'UTF8')), 'hex'), $2, 'queued')
ON CONFLICT (key) DO UPDATE SET job_id = EXCLUDED.job_id, state = 'queued', created_at = now(), done_at = NULL
WHERE ingest_job_keys.state = 'failed'
OR (ingest_job_keys.state = 'done' AND ingest_job_keys.done_at < now() - $3::interval)
OR (ingest_job_keys.state IN ('queued', 'running') AND ingest_job_keys.created_at < now() - $4::interval)
RETURNING key`, raw, id, KeyWindow, KeyStaleAfter).Scan(&key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("ingest job key: %w", err)
	}

	return true, nil
}

// settleKey marks the key of run id done or failed once the run ended.
func settleKey(ctx context.Context, ex database.Executor, id int64, status string) error {
	state := "failed"
	if status == StatusSucceeded || status == StatusPartial {
		state = "done"
	}

	_, err := ex.ExecContext(ctx, `UPDATE ingest_job_keys SET state = $2, done_at = now()
WHERE job_id = $1 AND state IN ('queued', 'running')`, id, state)

	return err
}

// Retry policy (RFC-001 section 4): five attempts, full jitter. workaround(vukem): VF-080, queue.Exponential has no
// jitter, so the handler registers this function through queue.Backoff.
const (
	RetryBase = 30 * time.Second
	RetryCap  = 30 * time.Minute
)

// JitterBackoff returns rand(0, min(cap, base*2^(attempt-1))) with attempt 1-based, as queue.Backoff expects. rnd
// yields a float in [0,1); tests pass a seeded source.
func JitterBackoff(base, limit time.Duration, rnd func() float64) func(attempt int) time.Duration {
	return func(attempt int) time.Duration {
		ceil := base
		for i := 1; i < attempt && ceil < limit; i++ {
			ceil *= 2
		}

		ceil = min(ceil, limit)

		return time.Duration(rnd() * float64(ceil))
	}
}

// Backoff is the production retry policy.
func Backoff() func(attempt int) time.Duration {
	return JitterBackoff(RetryBase, RetryCap, secureFloat)
}

// secureFloat is a uniform float in [0,1) from crypto/rand (the lint config bans math/rand); a failing reader
// yields 1, the longest wait, which is the safe side for a retry.
func secureFloat() float64 {
	const bits = 1 << 53

	n, err := rand.Int(rand.Reader, big.NewInt(bits))
	if err != nil {
		return 1
	}

	return float64(n.Int64()) / bits
}
