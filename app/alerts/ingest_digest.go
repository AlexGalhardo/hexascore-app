package alerts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// IngestWindow is how often the ingestion digest runs and how far back it looks (RFC-001 section 8).
const IngestWindow = 15 * time.Minute

// QuietAfter is how long without a new item turns the digest's first line into a warning.
const QuietAfter = 3 * IngestWindow

// LLMAlertUSD is the projected monthly LLM spend that gets flagged in the digest (RFC-001: US$18 of the US$20 ceiling).
const LLMAlertUSD = 18.0

// IngestError is one row of the "errors by source" line.
type IngestError struct {
	Source, Class string
	Count         int
}

// ActiveRun is one running job_runs row, for the "runs" lines.
type ActiveRun struct {
	Job, Source           string
	Age, BeatAge          time.Duration
	Seen, Changed, Failed int
}

// SourceCount is how many source records a source fetched in the window.
type SourceCount struct {
	Source string
	N      int
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
	// ErrorSample is the newest failure message in the window and its source and target; empty without failures.
	ErrorSample string
	// NewByType counts items created in the window; CreatedQuiet counts the last QuietAfter.
	NewByType    map[string]int
	CreatedQuiet int
	// CreatedHour counts the items created in the last hour: the pace and the ETA use it, a 15-minute window swung
	// from 900 to 2,800 items an hour between two digests.
	CreatedHour int
	Active      []ActiveRun
	Fetched     []SourceCount
	Games       int
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

	if err := each(ctx, db, `SELECT source, error_class, count(*) FROM ingest_failures
WHERE last_seen >= $1 AND last_seen < $2 GROUP BY source, error_class ORDER BY count(*) DESC, source, error_class
LIMIT 5`, []any{s.Since, at}, func(rows *sql.Rows) error {
		var e IngestError
		if err := rows.Scan(&e.Source, &e.Class, &e.Count); err != nil {
			return err
		}

		s.Errors = append(s.Errors, e)

		return nil
	}); err != nil {
		return s, fmt.Errorf("ingest digest: errors: %w", err)
	}

	if err := db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE type = 'book'), count(*) FILTER (WHERE type = 'movie'),
  count(*) FILTER (WHERE type = 'series'), count(*) FILTER (WHERE type = 'game') FROM items`).
		Scan(&s.Books, &s.Movies, &s.Series, &s.Games); err != nil {
		return s, fmt.Errorf("ingest digest: totals: %w", err)
	}

	if err := s.readProgress(ctx, db); err != nil {
		return s, err
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

	created := 0
	for _, n := range s.NewByType {
		created += n
	}

	fmt.Fprintf(
		&b,
		"<b>Ingest</b> %s (last %d min)\n",
		s.At.In(saoPaulo).Format("02/01 15:04"),
		int(IngestWindow.Minutes()),
	)

	total := s.Books + s.Movies + s.Series + s.Games

	switch {
	case created > 0:
		fmt.Fprintf(&b, "✅ +%d items · movies +%d · series +%d · books +%d · games +%d\n",
			created, s.NewByType["movie"], s.NewByType["series"], s.NewByType["book"], s.NewByType["game"])
	case s.CreatedQuiet == 0:
		fmt.Fprintf(&b, "⚠️ <b>0 new items in %d min</b> (%d runs active)\n", int(QuietAfter.Minutes()), len(s.Active))
	default:
		b.WriteString("⏸ 0 new items in this window\n")
	}

	fmt.Fprintf(&b, "Catalogue: %d of 1,000,000 (%.1f%%) · %d books · %d movies · %d series · %d games\n",
		total, float64(total)/10000, s.Books, s.Movies, s.Series, s.Games)

	if s.CreatedHour > 0 {
		perHour := float64(s.CreatedHour)
		fmt.Fprintf(
			&b,
			"Pace (last hour): %d items · ETA ≈ %.0f days\n",
			s.CreatedHour,
			float64(1_000_000-total)/perHour/24,
		)
	}

	fmt.Fprintf(&b, "\nRuns: %d running, %d queued\n", s.Running, s.Queued)

	for _, r := range s.Active {
		name := r.Job
		if r.Source != "" {
			name += "/" + r.Source
		}

		mark := "•"
		if r.BeatAge > 5*time.Minute {
			mark = "⚠️"
		}

		fmt.Fprintf(&b, "%s %s · %s · seen %s · changed %s · failed %d\n", mark, html.EscapeString(name),
			fmt.Sprintf("%dm", int(r.Age.Minutes())), compact(r.Seen), compact(r.Changed), r.Failed)
	}

	if len(s.Fetched) > 0 {
		parts := make([]string, len(s.Fetched))
		for i, c := range s.Fetched {
			parts[i] = fmt.Sprintf("%s %s", html.EscapeString(c.Source), compact(c.N))
		}

		fmt.Fprintf(
			&b,
			"Fetched: %s · rate limited %d · covers %d\n",
			strings.Join(parts, " · "),
			s.RateLimited,
			s.Covers,
		)
	}

	if len(s.Errors) > 0 {
		parts := make([]string, len(s.Errors))
		for i, e := range s.Errors {
			parts[i] = fmt.Sprintf("%s/%s ×%d", html.EscapeString(e.Source), html.EscapeString(e.Class), e.Count)
		}

		fmt.Fprintf(&b, "\nErrors: %s\n", strings.Join(parts, ", "))

		if s.ErrorSample != "" {
			fmt.Fprintf(&b, "  e.g. %s\n", html.EscapeString(s.ErrorSample))
		}
	}

	if s.HeartbeatOK {
		fmt.Fprintf(&b, "\nWorker: beat %s ago", s.HeartbeatAge.Round(time.Second))

		if s.HeartbeatAge > ingest.WorkerStaleAfter {
			b.WriteString(" <b>STALE</b>")
		}

		b.WriteString("\n")
	} else {
		b.WriteString("\nWorker: <b>no heartbeat</b>\n")
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

// each runs a query and hands every row to scan; the rows are closed and checked here, once.
func each(ctx context.Context, db database.Executor, query string, args []any, scan func(*sql.Rows) error) error {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// readProgress fills the fields that tell whether the catalogue is really growing.
func (s *IngestSummary) readProgress(ctx context.Context, db database.Executor) error {
	s.NewByType = map[string]int{}

	if err := each(ctx, db, `SELECT type, count(*) FROM items WHERE created_at >= $1 AND created_at < $2 GROUP BY type`,
		[]any{s.Since, s.At}, func(r *sql.Rows) error {
			var (
				t string
				n int
			)

			if err := r.Scan(&t, &n); err != nil {
				return err
			}

			s.NewByType[t] = n

			return nil
		}); err != nil {
		return fmt.Errorf("ingest digest: new by type: %w", err)
	}

	if err := db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE created_at >= $1), count(*) FROM items
WHERE created_at >= $2 AND created_at < $3`,
		s.At.Add(-QuietAfter), s.At.Add(-time.Hour), s.At).Scan(&s.CreatedQuiet, &s.CreatedHour); err != nil {
		return fmt.Errorf("ingest digest: quiet: %w", err)
	}

	if err := each(ctx, db, `SELECT job, coalesce(source, ''), seen, changed, failed,
  extract(epoch FROM $1::timestamptz - coalesce(started_at, created_at))::bigint,
  extract(epoch FROM $1::timestamptz - coalesce(heartbeat_at, started_at, created_at))::bigint
FROM job_runs WHERE status = 'running' ORDER BY started_at LIMIT 8`, []any{s.At}, func(r *sql.Rows) error {
		var (
			a         ActiveRun
			age, beat int64
		)

		if err := r.Scan(&a.Job, &a.Source, &a.Seen, &a.Changed, &a.Failed, &age, &beat); err != nil {
			return err
		}

		a.Age, a.BeatAge = time.Duration(age)*time.Second, time.Duration(beat)*time.Second
		s.Active = append(s.Active, a)

		return nil
	}); err != nil {
		return fmt.Errorf("ingest digest: active runs: %w", err)
	}

	if err := each(ctx, db, `SELECT source, count(*) FROM source_records WHERE fetched_at >= $1 AND fetched_at < $2
GROUP BY source ORDER BY count(*) DESC, source LIMIT 5`, []any{s.Since, s.At}, func(r *sql.Rows) error {
		var c SourceCount
		if err := r.Scan(&c.Source, &c.N); err != nil {
			return err
		}

		s.Fetched = append(s.Fetched, c)

		return nil
	}); err != nil {
		return fmt.Errorf("ingest digest: fetched: %w", err)
	}

	var source, target, msg string

	err := db.QueryRowContext(ctx, `SELECT source, external_id, left(error, 140) FROM ingest_failures
WHERE last_seen >= $1 AND last_seen < $2 ORDER BY last_seen DESC LIMIT 1`, s.Since, s.At).Scan(&source, &target, &msg)
	if err == nil {
		s.ErrorSample = fmt.Sprintf("%s %s: %s", source, target, msg)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("ingest digest: error sample: %w", err)
	}

	return nil
}

func compact(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}

	return strconv.Itoa(n)
}
