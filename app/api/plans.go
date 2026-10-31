package api

import (
	"context"
	"time"
)

// Plan IDs (SDD 05 FR-05-3).
const (
	PlanFree     = "free"
	PlanPro      = "pro"
	PlanBusiness = "business"
)

// Plan is a public API plan's limits, equal to legacy lib/api/plans.ts (D-33).
type Plan struct {
	ID string
	// Quota is requests per UTC calendar month, or over the account lifetime when Lifetime is set.
	Quota       int
	Lifetime    bool
	MaxPageSize int
	// RPS is the per-user throttle in requests per second.
	RPS int
	// Burst is how many requests may arrive at once; the sustained rate stays RPS (Q-A2).
	Burst int
}

// PlanByID returns the plan named id; unknown ids are not a plan.
func PlanByID(id string) (Plan, bool) {
	switch id {
	case PlanFree:
		return Plan{ID: PlanFree, Quota: 100, Lifetime: true, MaxPageSize: 10, RPS: 1, Burst: 5}, true
	case PlanPro:
		return Plan{ID: PlanPro, Quota: 50_000, MaxPageSize: 100, RPS: 10, Burst: 10}, true
	case PlanBusiness:
		return Plan{ID: PlanBusiness, Quota: 500_000, MaxPageSize: 100, RPS: 50, Burst: 50}, true
	}

	return Plan{}, false
}

// Period is the counter row the plan meters into at now: "life" or the UTC "YYYYMM" (SDD 05 §7.2).
func (p Plan) Period(now time.Time) string {
	if p.Lifetime {
		return "life"
	}

	return now.UTC().Format("200601")
}

// ResetIn is X-Quota-Reset: seconds until the next UTC month, 0 for a lifetime quota.
func (p Plan) ResetIn(now time.Time) int {
	if p.Lifetime {
		return 0
	}

	now = now.UTC()
	next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	return int(next.Sub(now).Seconds())
}

// PlanResolver returns a user's effective plan (SDD 05 §7.5).
type PlanResolver func(ctx context.Context, userID int64) (Plan, error)
