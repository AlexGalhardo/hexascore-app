package alerts

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// IngestWindow is how often the ingestion digest runs and how far back it looks (RFC-001 section 8).
const IngestWindow = 5 * time.Minute

// LLMAlertUSD is the projected monthly LLM spend that gets flagged in the digest (RFC-001: US$18 of the US$20 ceiling).
const LLMAlertUSD = 18.0

// IngestError is one row of the "errors by source" line.
type IngestError struct {
	Source, Class string
	Count         int
}

// IngestSummary is what the ingestion digest reports for [Since, At).
type IngestSummary struct {
	At, Since time.Time
	// Created counts items inserted in the window; Updated the records changed by runs that ended in it.
	Created, Updated, Failed, Covers, RateLimited int
	Queued, Running                               int
	// Books, Movies and Series are the catalogue totals; the 1,000,000-item run is followed through them.
	Books, Movies, Series int
	// HeartbeatAge is how long ago the worker last beat; HeartbeatOK is false when it never did.
	HeartbeatAge time.Duration
	HeartbeatOK  bool
	Errors       []IngestError
	// LLMMonthUSD is the month-to-date spend from ingest_cost_daily; LLMProjectedUSD extrapolates it to the month.
	LLMMonthUSD, LLMProjectedUSD float64
}

// BuildIngestSummary reads the tables for the window ending at at. It only reads.
func BuildIngestSummary(ctx context.Context, db database.Executor, at time.Time) (IngestSummary, error) {
	s := IngestSummary{At: at, Since: at.Add(-IngestWindow)}

	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM items WHERE created_at >= $1 AND created_at < $2`,
		s.Since, at).Scan(&s.Created); err != nil {
		return s, fmt.Errorf("ingest digest: created: %w", err)
	}

	if err := db.QueryRowContext(ctx, `SELECT
  coalesce(sum(changed) FILTER (WHERE job <> 'covers'), 0), coalesce(sum(failed), 0),
  coalesce(sum(changed) FILTER (WHERE job = 'covers'), 0), coalesce(sum(rate_limited), 0)
FROM job_runs WHERE finished_at >= $1 AND finished_at < $2`, s.Since, at).
		Scan(&s.Updated, &s.Failed, &s.Covers, &s.RateLimited); err != nil {
		return s, fmt.Errorf("ingest digest: runs: %w", err)
	}

	if err := db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE status = 'queued'),
  count(*) FILTER (WHERE status = 'running') FROM job_runs WHERE status IN ('queued', 'running')`).
		Scan(&s.Queued, &s.Running); err != nil {
		return s, fmt.Errorf("ingest digest: backlog: %w", err)
	}

	rows, err := db.QueryContext(ctx, `SELECT source, error_class, count(*) FROM ingest_failures
WHERE last_seen >= $1 AND last_seen < $2 GROUP BY source, error_class ORDER BY count(*) DESC, source, error_class
LIMIT 5`, s.Since, at)
	if err != nil {
		return s, fmt.Errorf("ingest digest: errors: %w", err)
	}

	for rows.Next() {
		var e IngestError
		if err := rows.Scan(&e.Source, &e.Class, &e.Count); err != nil {
			_ = rows.Close()

			return s, err
		}

		s.Errors = append(s.Errors, e)
	}

	if err := rows.Err(); err != nil {
		_ = rows.Close()

		return s, err
	}

	if err := rows.Close(); err != nil {
		return s, err
	}

	if err := db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE type = 'book'), count(*) FILTER (WHERE type = 'movie'),
  count(*) FILTER (WHERE type = 'series') FROM items`).
		Scan(&s.Books, &s.Movies, &s.Series); err != nil {
		return s, fmt.Errorf("ingest digest: totals: %w", err)
	}

	if beat, ok, err := ingest.ReadWorkerHeartbeat(ctx, db); err != nil {
		return s, err
	} else if ok {
		s.HeartbeatOK, s.HeartbeatAge = true, at.Sub(beat)
	}

	day := at.In(saoPaulo)

	if err := db.QueryRowContext(ctx, `SELECT coalesce(sum(llm_usd), 0)::float8 FROM ingest_cost_daily
WHERE day >= date_trunc('month', $1::date) AND day <= $1::date`, day.Format(time.DateOnly)).
		Scan(&s.LLMMonthUSD); err != nil {
		return s, fmt.Errorf("ingest digest: llm spend: %w", err)
	}

	daysInMonth := time.Date(day.Year(), day.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	s.LLMProjectedUSD = s.LLMMonthUSD / float64(day.Day()) * float64(daysInMonth)

	return s, nil
}

// FormatIngest renders the digest as the HTML subset Telegram accepts. The time is America/Sao_Paulo.
func FormatIngest(s IngestSummary) string {
	var b strings.Builder

	fmt.Fprintf(
		&b,
		"<b>Ingest digest</b> %s (last %d min)\n",
		s.At.In(saoPaulo).Format("02/01 15:04"),
		int(IngestWindow.Minutes()),
	)
	fmt.Fprintf(&b, "Created %d · updated %d · failed %d · covers %d · rate limited %d\n",
		s.Created, s.Updated, s.Failed, s.Covers, s.RateLimited)
	fmt.Fprintf(&b, "Backlog: %d queued, %d running\n", s.Queued, s.Running)
	fmt.Fprintf(&b, "Catalogue: %d books · %d movies · %d series (%d of 1,000,000)\n",
		s.Books, s.Movies, s.Series, s.Books+s.Movies+s.Series)

	if s.HeartbeatOK {
		fmt.Fprintf(&b, "Worker: beat %s ago", s.HeartbeatAge.Round(time.Second))

		if s.HeartbeatAge > ingest.WorkerStaleAfter {
			b.WriteString(" <b>STALE</b>")
		}

		b.WriteString("\n")
	} else {
		b.WriteString("Worker: <b>no heartbeat</b>\n")
	}

	if len(s.Errors) > 0 {
		parts := make([]string, len(s.Errors))
		for i, e := range s.Errors {
			parts[i] = fmt.Sprintf("%s/%s ×%d", html.EscapeString(e.Source), html.EscapeString(e.Class), e.Count)
		}

		fmt.Fprintf(&b, "Errors: %s\n", strings.Join(parts, ", "))
	}

	fmt.Fprintf(
		&b,
		"LLM: $%.2f this month, projected $%.2f (alert at $%.2f)",
		s.LLMMonthUSD,
		s.LLMProjectedUSD,
		LLMAlertUSD,
	)

	if s.LLMProjectedUSD >= LLMAlertUSD {
		b.WriteString(" <b>over the alert</b>")
	}

	return b.String()
}

// SendIngestDigest posts the digest to the ops chat every time it runs, quiet or not: while the catalogue is being
// filled, a missing message is the signal that something stopped. With a nil notifier (Telegram unset) it sends
// nothing. It reports whether a message went out.
func SendIngestDigest(ctx context.Context, db database.Executor, n *Notifier, at time.Time, _ bool) (bool, error) {
	if n == nil {
		return false, nil
	}

	s, err := BuildIngestSummary(ctx, db, at)
	if err != nil {
		return false, err
	}

	if err := n.SendOps(ctx, FormatIngest(s)); err != nil {
		return false, err
	}

	return true, nil
}
