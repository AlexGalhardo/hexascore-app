package llm

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// Prices of Claude Haiku 5.5 in US dollars per million tokens, for prompts up to 100K tokens (RFC-001 section 6, read on
// the provider's pricing page on 2026-11-02; re-verify before raising the ceiling).
const (
	inputPerMTok  = 0.10
	outputPerMTok = 0.50
	// perSearch is the web search price, US$10 per 1,000 searches (re-verify on the pricing page before relying on it).
	perSearch = 0.01
)

// DefaultBudgetUSD is the monthly ceiling when LLM_MONTHLY_BUDGET_USD is unset or unusable (owner 2026-11-04: US$10).
const DefaultBudgetUSD = 10.0

// Cost is the price in US dollars of a call without web search.
func Cost(tokensIn, tokensOut int) float64 { return CostWithSearches(tokensIn, tokensOut, 0) }

// CostWithSearches is the price in US dollars of a call that ran searches web searches.
func CostWithSearches(tokensIn, tokensOut, searches int) float64 {
	return float64(tokensIn)*inputPerMTok/1e6 + float64(tokensOut)*outputPerMTok/1e6 + float64(searches)*perSearch
}

// Ceiling reads LLM_MONTHLY_BUDGET_USD through get. Unset, negative or not a finite number gives the default; 0 is a
// valid ceiling and stops every call.
func Ceiling(get func(string) string) float64 {
	v, err := strconv.ParseFloat(get("LLM_MONTHLY_BUDGET_USD"), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return DefaultBudgetUSD
	}

	return v
}

// Spent is the cost recorded in llm_usage for the UTC month of now.
func Spent(ctx context.Context, ex database.Executor, now time.Time) (float64, error) {
	start := time.Date(now.UTC().Year(), now.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)

	var spent sql.NullFloat64

	err := ex.QueryRowContext(ctx, `SELECT sum(cost_usd)::float8 FROM llm_usage WHERE ts >= $1 AND ts < $2`,
		start, start.AddDate(0, 1, 0)).Scan(&spent)
	if err != nil {
		return 0, fmt.Errorf("llm: month spend: %w", err)
	}

	return spent.Float64, nil
}

// DailyShare spreads the monthly ceiling: a day may spend this fraction of it, so the backlog is worked a little every
// day instead of the whole budget going on the first night (owner 2026-11-04: US$10 a month).
const DailyShare = 1.0 / 30

// SpentToday is the cost recorded in llm_usage since the start of the UTC day of now.
func SpentToday(ctx context.Context, ex database.Executor, now time.Time) (float64, error) {
	start := now.UTC().Truncate(24 * time.Hour)

	var spent sql.NullFloat64

	err := ex.QueryRowContext(ctx, `SELECT sum(cost_usd)::float8 FROM llm_usage WHERE ts >= $1`, start).Scan(&spent)
	if err != nil {
		return 0, fmt.Errorf("llm: day spend: %w", err)
	}

	return spent.Float64, nil
}

// dayDone reports whether today's share of the ceiling is spent: the run then ends quietly and the next day goes on.
func dayDone(ctx context.Context, ex database.Executor, now time.Time, ceiling float64) (bool, error) {
	today, err := SpentToday(ctx, ex, now)
	if err != nil {
		return false, err
	}

	return today >= ceiling*DailyShare, nil
}

// Usage is one row of llm_usage.
type Usage struct {
	Job, ItemID, Outcome string
	TokensIn, TokensOut  int
	Searches             int
}

// Record stores a call and its cost. Outcome says what became of the answer: queued, empty or invalid.
func Record(ctx context.Context, ex database.Executor, u Usage, now time.Time) error {
	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO llm_usage (ts, job, model, item_id, tokens_in, tokens_out, cost_usd, outcome)
VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6, $7, $8)`,
		now,
		u.Job,
		Model,
		u.ItemID,
		u.TokensIn,
		u.TokensOut,
		CostWithSearches(u.TokensIn, u.TokensOut, u.Searches),
		u.Outcome,
	)
	if err != nil {
		return fmt.Errorf("llm: record usage: %w", err)
	}

	return nil
}
