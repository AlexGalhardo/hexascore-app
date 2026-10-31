package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/httpclient"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const honeypotWindow = 7 * 24 * time.Hour

// ErrCoverBlocked is a cover URL the safe client will not fetch.
var ErrCoverBlocked = errors.New("community: cover url blocked")

// ErrCoverRefused is a cover URL or body that fails the cover rules.
// It is terminal: the job records it and does not retry.
var ErrCoverRefused = errors.New("community: cover refused")

// Burst puts an IP limiter in front of a per-user limiter (SDD 04 §7.3).
// An anonymous caller falls through ByUser onto the IP key.
func Burst(ipLimit, userLimit ratelimit.Limiter, guard string) func(http.Handler) http.Handler {
	perUser := ratelimit.Middleware(userLimit, ratelimit.ByUser(guard))

	return func(next http.Handler) http.Handler {
		return ratelimit.Middleware(ipLimit, ratelimit.ByIP)(perUser(next))
	}
}

// DraftBurst is 30 saves an hour, per IP and then per user.
func DraftBurst(f ratelimit.Factory, guard string) func(http.Handler) http.Handler {
	return Burst(limits.CommunityDraftIP.From(f), limits.CommunityDraft.From(f), guard)
}

// SubmitBurst is 10 submits or withdrawals a minute, per IP and then per user.
func SubmitBurst(f ratelimit.Factory, guard string) func(http.Handler) http.Handler {
	return Burst(limits.CommunitySubmitIP.From(f), limits.CommunitySubmit.From(f), guard)
}

// UploadBurst is 10 cover uploads an hour, per IP and then per user.
func UploadBurst(f ratelimit.Factory, guard string) func(http.Handler) http.Handler {
	return Burst(limits.CommunityUploadIP.From(f), limits.CommunityUpload.From(f), guard)
}

// RejectHoneypot answers a filled website field with a generic 422 and logs
// community.honeypot. The field value is not logged. An empty field is not a
// hit. The third hit in seven days sets tier blocked for 30 days. db may be
// nil when the caller only needs the refusal and the log line.
func RejectHoneypot(
	ctx context.Context,
	db *database.DB,
	r *http.Request,
	userID int64,
	website string,
	now time.Time,
) error {
	if strings.TrimSpace(website) == "" {
		return nil
	}

	if r != nil {
		web.SecurityEvent(r, "community.honeypot", "user_id", userID)
	}

	if db != nil && userID != 0 {
		if err := noteHoneypot(ctx, db, userID, now); err != nil {
			return err
		}
	}

	return honeypotProblem()
}

func honeypotProblem() *Problem {
	return &Problem{
		Status: http.StatusUnprocessableEntity,
		Code:   "validation",
		Type:   problemBase + "validation",
	}
}

func noteHoneypot(ctx context.Context, db *database.DB, userID int64, now time.Time) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO community_honeypot_hits (user_id, created_at) VALUES ($1, $2)`,
			userID,
			now.UTC(),
		); err != nil {
			return fmt.Errorf("community: record honeypot: %w", err)
		}

		var n int
		err := tx.QueryRowContext(
			ctx,
			`SELECT count(*) FROM community_honeypot_hits WHERE user_id = $1 AND created_at >= $2`,
			userID,
			now.UTC().Add(-honeypotWindow),
		).Scan(&n)
		if err != nil {
			return fmt.Errorf("community: count honeypot: %w", err)
		}

		if n < 3 {
			return nil
		}

		until := now.UTC().Add(blockFor)
		if _, err := tx.ExecContext(ctx, `INSERT INTO contributor_stats (user_id, tier, blocked_until, updated_at)
VALUES ($1, 'blocked', $2, $3)
ON CONFLICT (user_id) DO UPDATE SET tier = 'blocked', blocked_until = EXCLUDED.blocked_until, updated_at = EXCLUDED.updated_at`,
			userID, until, now.UTC()); err != nil {
			return fmt.Errorf("community: block honeypot: %w", err)
		}

		return nil
	})
}

// SourceClient is the safe client for a cited URL: 5s, 3 redirects, 256 KiB.
func SourceClient() *http.Client {
	return httpclient.Safe(httpclient.Options{
		Timeout:      5 * time.Second,
		MaxRedirects: 3,
		MaxBody:      256 << 10,
	})
}

// SourceStatus fetches one cited URL. A bad URL is refused or blocked and is
// not a job failure: the check never blocks the submission.
func SourceStatus(ctx context.Context, raw string) string {
	if err := allowedSourceURL(raw); err != nil {
		return "refused"
	}

	client := SourceClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, http.NoBody)
	if err != nil {
		return "refused"
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, httpclient.ErrBlockedAddress) {
			return "blocked"
		}

		return "unreachable"
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil || resp.StatusCode >= 400 {
		return "unreachable"
	}

	return "ok"
}

// CheckSources stores source_status in meta. A missing suggestion is a no-op.
// Probe failures stay in the status map; they do not fail the job.
func CheckSources(ctx context.Context, db *database.DB, suggestionID string) error {
	var urls models.TextArray
	err := db.QueryRowContext(ctx, `SELECT source_urls FROM suggestions WHERE id = $1`, suggestionID).Scan(&urls)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("community: load source urls: %w", err)
	}

	status := make(map[string]string, len(urls))
	for _, raw := range urls {
		status[raw] = SourceStatus(ctx, raw)
	}

	body, err := json.Marshal(map[string]any{"source_status": status})
	if err != nil {
		return fmt.Errorf("community: encode source status: %w", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE suggestions SET meta = meta || $2::jsonb, updated_at = now() WHERE id = $1`,
		suggestionID,
		body,
	); err != nil {
		return fmt.Errorf("community: store source status: %w", err)
	}

	return nil
}
