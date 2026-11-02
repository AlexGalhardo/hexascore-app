package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
)

// ingest_host_state rows (RFC-001 section 3): key "web" holds the pressure Signal the web process writes, key
// "worker" the worker heartbeat, key "host:<name>" the persisted httpx.HostState of an outbound host.
const (
	KeyWeb     = "web"
	hostPrefix = "host:"
)

// StaleAfter is the age after which a pressure signal says nothing: the web process is down or not writing.
const StaleAfter = time.Minute

// Signal is what the web process publishes about its own load.
type Signal struct {
	InFlight int       `json:"inflight"`
	P95MS    int64     `json:"p95_ms"`
	At       time.Time `json:"at"`
}

// Busy applies the SDD 06 section 7.5 thresholds.
func (s Signal) Busy() bool {
	return s.InFlight > MaxInFlight || time.Duration(s.P95MS)*time.Millisecond > MaxP95
}

// Signal snapshots the gauges.
func (p *Pressure) Signal() Signal {
	return Signal{InFlight: p.InFlight(), P95MS: p.P95().Milliseconds(), At: p.now()}
}

func putState(ctx context.Context, ex database.Executor, key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = ex.ExecContext(ctx, `INSERT INTO ingest_host_state (key, data, updated_at) VALUES ($1, $2, now())
ON CONFLICT (key) DO UPDATE SET data = EXCLUDED.data, updated_at = now()`, key, b)
	if err != nil {
		return fmt.Errorf("ingest host state %s: %w", key, err)
	}

	return nil
}

// WriteSignal stores the web process's pressure signal (one UPSERT).
func WriteSignal(ctx context.Context, ex database.Executor, s Signal) error {
	return putState(ctx, ex, KeyWeb, s)
}

// ReadSignal returns the stored signal; ok is false when none exists.
func ReadSignal(ctx context.Context, ex database.Executor) (s Signal, ok bool, err error) {
	var b []byte

	err = ex.QueryRowContext(ctx, `SELECT data FROM ingest_host_state WHERE key = $1`, KeyWeb).Scan(&b)
	if errors.Is(err, sql.ErrNoRows) {
		return Signal{}, false, nil
	}

	if err != nil {
		return Signal{}, false, fmt.Errorf("ingest pressure signal: %w", err)
	}

	if err := json.Unmarshal(b, &s); err != nil {
		return Signal{}, false, fmt.Errorf("ingest pressure signal: %w", err)
	}

	return s, true, nil
}

// SignalGate holds a batch while the signal in the table says the site is busy. A missing or stale signal
// (older than StaleAfter) means "unknown" and lets the batch go; Local, when set, is also consulted so the web
// process keeps its immediate gauges. Like Pressure.Wait it fails open after MaxWait.
type SignalGate struct {
	DB    database.Executor
	Local *Pressure
	Now   func() time.Time
	// Sleep replaces the real wait; tests only.
	Sleep func(ctx context.Context, d time.Duration) error
}

func (g *SignalGate) busy(ctx context.Context) (bool, error) {
	if g.Local != nil && g.Local.busy() {
		return true, nil
	}

	s, ok, err := ReadSignal(ctx, g.DB)
	if err != nil {
		return false, err
	}

	return ok && g.Now().Sub(s.At) <= StaleAfter && s.Busy(), nil
}

// Wait returns when the site is quiet, after MaxWait, or when ctx ends. A failed read does not stall ingestion:
// the batch goes on (the error is returned only for a cancelled context).
func (g *SignalGate) Wait(ctx context.Context) error {
	sleepFn := g.Sleep
	if sleepFn == nil {
		sleepFn = sleep
	}

	deadline := g.Now().Add(MaxWait)

	for g.Now().Before(deadline) {
		busy, err := g.busy(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil || !busy {
			return nil //nolint:nilerr // a failed read must not stall ingestion
		}

		if err := sleepFn(ctx, PollEvery); err != nil {
			return err
		}
	}

	return nil
}

// HostStore persists httpx.HostState in ingest_host_state; it implements httpx.StateStore.
type HostStore struct{ DB database.Executor }

var _ httpx.StateStore = HostStore{}

// Load returns the stored state of every host.
func (s HostStore) Load(ctx context.Context) (map[string]httpx.HostState, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT key, data FROM ingest_host_state WHERE key LIKE $1`, hostPrefix+"%")
	if err != nil {
		return nil, fmt.Errorf("ingest host state: %w", err)
	}
	defer rows.Close()

	out := map[string]httpx.HostState{}

	for rows.Next() {
		var (
			key string
			b   []byte
			st  httpx.HostState
		)

		if err := rows.Scan(&key, &b); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &st); err != nil {
			return nil, fmt.Errorf("ingest host state %s: %w", key, err)
		}

		out[strings.TrimPrefix(key, hostPrefix)] = st
	}

	return out, rows.Err()
}

// Save stores one host's state.
func (s HostStore) Save(ctx context.Context, host string, st httpx.HostState) error {
	return putState(ctx, s.DB, hostPrefix+host, st)
}
