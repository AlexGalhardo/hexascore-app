// Package ingest holds the connector contract of the ingestion pipeline
// (SDD 06 §7.3). Connectors live in sub-packages.
package ingest

import (
	"context"
	"iter"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// Record is one fetched source payload.
type Record = store.Record

// Client is the outbound client shared by every connector.
type Client = httpx.Client

// Target is one thing a connector fetches: a Wikidata QID with its item
// type, or a Wikipedia title key.
type Target struct {
	ID   string
	Type string
}

// RunParams is the subset of job_runs.params a connector reads.
type RunParams struct {
	// RunID lets a connector cache per-run work; zero disables caching.
	RunID int64
	Mode  string
	Since time.Time
	Types []string
	Limit int
}

// TargetError ties a failed target to its cause for the dead-letter table.
type TargetError struct {
	Target Target
	Err    error
}

// Connector fetches one source. Targets yields work after the cursor
// position `after` (empty starts from the beginning); Fetch resolves one
// batch.
type Connector interface {
	Source() string
	Targets(ctx context.Context, db database.Executor, p RunParams, after string) iter.Seq2[Target, error]
	Fetch(ctx context.Context, c *Client, batch []Target) ([]Record, []TargetError)
}
