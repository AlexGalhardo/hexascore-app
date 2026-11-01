// Package dq runs the data-quality asserts of the catalogue (SDD 01 §7.6 DQ-01..14, SDD 06 §7.7 DQ-I-01..13) and
// writes one dq_results row per code.
package dq

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// Job is the job_runs.job of the asserts.
const Job = "dq.asserts"

// SampleLimit is the most offending ids kept in a result.
const SampleLimit = 20

// Count drops over this share of the last snapshot fail DQ-I-01.
const maxCountDrop = 0.01

//go:embed sql/*.sql
var sqlFiles embed.FS

// check is one assert backed by a SQL file that selects the offending ids.
type check struct {
	// Code is the dq_results code; Also lists the SDD 06 codes that restate the same assert.
	Code     string
	Also     []string
	Severity string
	// Max is the offender count that still passes.
	Max int
	// Report asserts always pass: the rows are listed for review.
	Report bool
}

var checks = []check{
	{Code: "DQ-01", Also: []string{"DQ-I-02"}, Severity: "error"},
	{Code: "DQ-02", Also: []string{"DQ-I-03"}, Severity: "error"},
	{Code: "DQ-03", Also: []string{"DQ-I-07"}, Severity: "warn", Report: true},
	{Code: "DQ-04", Also: []string{"DQ-I-08"}, Severity: "error"},
	{Code: "DQ-05", Severity: "error"},
	{Code: "DQ-06", Also: []string{"DQ-I-05"}, Severity: "error"},
	{Code: "DQ-07", Severity: "error"},
	{Code: "DQ-08", Severity: "error"},
	{Code: "DQ-09", Severity: "warn", Report: true},
	{Code: "DQ-10", Severity: "warn"},
	{Code: "DQ-11", Severity: "error"},
	{Code: "DQ-12", Severity: "warn"},
	{Code: "DQ-13", Severity: "warn"},
	{Code: "DQ-14", Severity: "warn"},
	{Code: "DQ-I-06", Severity: "error"},
	{Code: "DQ-I-09", Severity: "error"},
	{Code: "DQ-I-10", Severity: "warn", Report: true},
	{Code: "DQ-I-11", Severity: "warn", Max: 500},
	{Code: "DQ-I-13", Severity: "error"},
}

// Codes lists every code a run writes (without the optional DQ-I-04 and DQ-I-12), sorted.
func Codes() []string {
	var out []string

	for _, c := range checks {
		out = append(out, c.Code)
		out = append(out, c.Also...)
	}

	out = append(out, "DQ-I-01")
	slices.Sort(out)

	return out
}

// Stat reports whether the object stored under key exists (DQ-I-04).
type Stat func(ctx context.Context, key string) (bool, error)

// Serialiser returns the ids of items whose public API payload leaks a display-only source (DQ-I-12).
type Serialiser func(ctx context.Context, limit int) ([]string, error)

// Asserts is the dq.asserts step. Stat and Serialise are optional: their asserts are skipped while they are nil
// (the covers store arrives with P6.B.2.2 and the serialiser probe with the API wiring).
type Asserts struct {
	Stat      Stat
	Serialise Serialiser
	// CoverSample is how many covers DQ-I-04 stats in a daily run; zero means 1000. A full run stats all of them.
	CoverSample int
}

type result struct {
	code, severity      string
	passed              bool
	observed, threshold float64
	sample              []string
	hasThreshold        bool
}

// Step runs every assert once and writes the results for the run.
func (a Asserts) Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	var results []result

	for _, c := range checks {
		q, err := sqlFiles.ReadFile("sql/" + c.Code + ".sql")
		if err != nil {
			return run.Result{}, fmt.Errorf("dq: %s: %w", c.Code, err)
		}

		n, sample, err := offenders(ctx, tx, string(q))
		if err != nil {
			return run.Result{}, fmt.Errorf("dq: %s: %w", c.Code, err)
		}

		for _, code := range append([]string{c.Code}, c.Also...) {
			results = append(results, result{
				code: code, severity: c.Severity, passed: c.Report || n <= c.Max,
				observed: float64(n), threshold: float64(c.Max), sample: sample, hasThreshold: !c.Report,
			})
		}
	}

	drop, err := a.countDrop(ctx, tx, r)
	if err != nil {
		return run.Result{}, err
	}

	results = append(results, drop)

	optional, err := a.optional(ctx, tx, r)
	if err != nil {
		return run.Result{}, err
	}

	results = append(results, optional...)

	failed := 0

	for _, x := range results {
		if err := write(ctx, tx, r.ID, x); err != nil {
			return run.Result{}, err
		}

		if !x.passed && x.severity == "error" {
			failed++
		}
	}

	return run.Result{Done: true, Seen: len(results), Targets: len(results), Failed: failed}, nil
}

// offenders counts the rows of q (a query of ids) and keeps the first SampleLimit.
func offenders(ctx context.Context, tx *database.Tx, q string) (int, []string, error) {
	var (
		n   int
		raw []byte
	)

	err := tx.QueryRowContext(ctx, `WITH bad(id) AS (`+q+`)
SELECT (SELECT count(*) FROM bad),
  coalesce((SELECT jsonb_agg(id) FROM (SELECT id FROM bad ORDER BY id LIMIT $1) s), '[]')`, SampleLimit).
		Scan(&n, &raw)
	if err != nil {
		return 0, nil, err
	}

	var sample []string

	return n, sample, json.Unmarshal(raw, &sample)
}

func write(ctx context.Context, tx *database.Tx, runID int64, x result) error {
	sample, err := json.Marshal(append([]string{}, x.sample...))
	if err != nil {
		return err
	}

	var threshold any
	if x.hasThreshold {
		threshold = x.threshold
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO dq_results (run_id, code, severity, passed, observed, threshold, sample)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
ON CONFLICT (run_id, code) DO UPDATE SET severity = excluded.severity, passed = excluded.passed,
observed = excluded.observed, threshold = excluded.threshold, sample = excluded.sample, checked_at = now()`,
		runID, x.code, x.severity, x.passed, x.observed, threshold, string(sample))
	if err != nil {
		return fmt.Errorf("dq: write %s: %w", x.code, err)
	}

	return nil
}
