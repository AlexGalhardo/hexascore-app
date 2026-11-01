package alerts

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/web"
)

// FlushInterval is how often the recorder writes its counts: one upsert per fingerprint and hour, never per error.
const FlushInterval = time.Minute

// digestTop is how many fingerprints the daily digest lists (SDD 12 FR-12-6).
const digestTop = 20

// digestRetention is how long error_digest_daily keeps its rows.
const digestRetention = 90 * 24 * time.Hour

// maxPending bounds the fingerprints buffered between flushes; new ones past it are dropped (the log lines stay in
// Railway) instead of growing memory.
const maxPending = 1000

// maxMessage bounds the stored message, so 20 lines fit in one Telegram message.
const maxMessage = 120

type pendingKey struct {
	hour time.Time
	fp   string
}

type pendingRow struct {
	message, alert, route, requestID string
	count                            int64
}

// Recorder is an slog.Handler that counts ERROR records and alert lines per fingerprint and hour in memory and
// writes them to error_digest_daily every FlushInterval (SDD 12 §7.2). Install it after the redacting log handler,
// so stored messages are already scrubbed.
type Recorder struct {
	db  database.Executor
	now func() time.Time

	mu      sync.Mutex
	pending map[pendingKey]*pendingRow
	attrs   []slog.Attr
	shared  *Recorder
}

// NewRecorder returns a recorder writing to db.
func NewRecorder(db database.Executor, now func() time.Time) *Recorder {
	r := &Recorder{db: db, now: now, pending: map[pendingKey]*pendingRow{}}
	r.shared = r

	return r
}

// Enabled implements slog.Handler.
func (r *Recorder) Enabled(_ context.Context, l slog.Level) bool { return l >= slog.LevelWarn }

// Handle implements slog.Handler.
func (r *Recorder) Handle(ctx context.Context, rec slog.Record) error {
	f := Fields(r.attrs, rec)
	if rec.Level < slog.LevelError && !alertCode.MatchString(f["alert"]) {
		return nil
	}

	if f["request_id"] == "" {
		f["request_id"] = web.RequestID(ctx)
	}

	k := pendingKey{hour: r.now().UTC().Truncate(time.Hour), fp: Fingerprint(rec.Message, f)}

	s := r.shared
	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.pending[k]
	if !ok {
		if len(s.pending) >= maxPending {
			return nil
		}

		row = &pendingRow{message: truncate(rec.Message), alert: f["alert"], route: f["route"]}
		s.pending[k] = row
	}

	row.count++
	if row.requestID == "" {
		row.requestID = f["request_id"]
	}

	return nil
}

// WithAttrs implements slog.Handler.
func (r *Recorder) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Recorder{
		db:     r.db,
		now:    r.now,
		shared: r.shared,
		attrs:  append(append([]slog.Attr(nil), r.attrs...), attrs...),
	}
}

// WithGroup implements slog.Handler; grouped attributes do not identify an error.
func (r *Recorder) WithGroup(string) slog.Handler { return r }

// Flush writes the buffered counts. Rows that fail stay buffered for the next flush.
func (r *Recorder) Flush(ctx context.Context) error {
	s := r.shared
	s.mu.Lock()
	batch := s.pending
	s.pending = map[pendingKey]*pendingRow{}
	s.mu.Unlock()

	var errs []error

	for k, row := range batch {
		_, err := s.db.ExecContext(ctx, `INSERT INTO error_digest_daily (hour, fingerprint, message, alert, route,
  request_id, count)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (hour, fingerprint) DO UPDATE SET count = error_digest_daily.count + EXCLUDED.count,
  request_id = coalesce(nullif(error_digest_daily.request_id, ''), EXCLUDED.request_id)`,
			k.hour, k.fp, row.message, row.alert, row.route, row.requestID, row.count)
		if err != nil {
			errs = append(errs, err)

			s.mu.Lock()
			if cur, ok := s.pending[k]; ok {
				cur.count += row.count
			} else {
				s.pending[k] = row
			}
			s.mu.Unlock()
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error digest flush: %w", errors.Join(errs...))
	}

	return nil
}

// Run flushes every FlushInterval until ctx ends, then once more so shutdown keeps the last interval (vukem.Runner).
func (r *Recorder) Run(ctx context.Context) error {
	tick := time.NewTicker(FlushInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// A failed flush is retried next tick; logging it at ERROR would feed the recorder itself.
			_ = r.Flush(ctx)
		case <-ctx.Done():
			final, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()

			return r.Flush(final)
		}
	}
}

// Digest builds the daily summary of the 24 h before now: totals, alert codes, failed job runs and the top 20
// fingerprints with one request id each, as Telegram HTML.
func Digest(ctx context.Context, db database.Executor, now time.Time) (string, error) {
	since := now.Add(-24 * time.Hour).UTC().Truncate(time.Hour)

	var total, distinct int64
	if err := db.QueryRowContext(ctx, `SELECT coalesce(sum(count), 0), count(DISTINCT fingerprint)
FROM error_digest_daily WHERE hour >= $1`, since).Scan(&total, &distinct); err != nil {
		return "", fmt.Errorf("digest totals: %w", err)
	}

	var failedRuns int64
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM job_runs WHERE status = 'failed' AND finished_at > $1`,
		now.Add(-24*time.Hour)).Scan(&failedRuns); err != nil {
		return "", fmt.Errorf("digest job runs: %w", err)
	}

	var b strings.Builder

	b.WriteString("<b>Daily error digest</b> (last 24 h)\n")

	if total == 0 && failedRuns == 0 {
		b.WriteString("No errors recorded.")
		return b.String(), nil
	}

	fmt.Fprintf(&b, "Errors: %d in %d fingerprints. Failed job runs: %d\n", total, distinct, failedRuns)

	alerts, err := db.QueryContext(ctx, `SELECT alert, sum(count) FROM error_digest_daily
WHERE hour >= $1 AND alert <> '' GROUP BY alert ORDER BY alert`, since)
	if err != nil {
		return "", fmt.Errorf("digest alerts: %w", err)
	}

	var codes []string

	for alerts.Next() {
		var code string

		var n int64
		if err := alerts.Scan(&code, &n); err != nil {
			_ = alerts.Close()
			return "", err
		}

		codes = append(codes, fmt.Sprintf("%s: %d", html.EscapeString(code), n))
	}

	if err := errors.Join(alerts.Err(), alerts.Close()); err != nil {
		return "", err
	}

	if len(codes) > 0 {
		b.WriteString("Alerts: " + strings.Join(codes, ", ") + "\n")
	}

	rows, err := db.QueryContext(ctx, `SELECT max(message), max(route), sum(count), max(request_id)
FROM error_digest_daily WHERE hour >= $1
GROUP BY fingerprint ORDER BY sum(count) DESC, fingerprint LIMIT $2`, since, digestTop)
	if err != nil {
		return "", fmt.Errorf("digest top: %w", err)
	}

	b.WriteString("\n")

	for i := 1; rows.Next(); i++ {
		var msg, route, rid string

		var n int64
		if err := rows.Scan(&msg, &route, &n, &rid); err != nil {
			_ = rows.Close()
			return "", err
		}

		fmt.Fprintf(&b, "%d. %d× <code>%s</code>", i, n, html.EscapeString(msg))

		if route != "" {
			fmt.Fprintf(&b, " %s", html.EscapeString(route))
		}

		if rid != "" {
			fmt.Fprintf(&b, " rid %s", html.EscapeString(rid))
		}

		b.WriteString("\n")
	}

	return b.String(), errors.Join(rows.Err(), rows.Close())
}

// SendDigest prunes rows past the retention and posts the digest to the error chat; with a nil notifier (Telegram
// unset) it only prunes.
func SendDigest(ctx context.Context, db database.Executor, n *Notifier, now time.Time) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM error_digest_daily WHERE hour < $1`,
		now.Add(-digestRetention)); err != nil {
		return fmt.Errorf("digest prune: %w", err)
	}

	if n == nil {
		return nil
	}

	text, err := Digest(ctx, db, now)
	if err != nil {
		return err
	}

	return n.SendSilent(ctx, text)
}

func truncate(s string) string {
	if utf8.RuneCountInString(s) <= maxMessage {
		return s
	}

	return string([]rune(s)[:maxMessage-1]) + "…"
}
