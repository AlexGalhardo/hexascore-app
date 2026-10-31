package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// slaTarget is the NFR-04-2 decision budget (SDD 04): a decision later than this counts as an SLA breach.
const slaTarget = 24 * time.Hour

const openSuggestions = ` FROM suggestions WHERE status IN ('submitted','in_review')`

var communityGauges = []struct{ name, sql string }{
	{"suggest.queue.size", `SELECT count(*)::float8` + openSuggestions},
	{"suggest.queue.age_p50_seconds", `SELECT COALESCE(percentile_cont(0.5) WITHIN GROUP
  (ORDER BY EXTRACT(EPOCH FROM now() - submitted_at)), 0)::float8` + openSuggestions},
	{"suggest.queue.age_p95_seconds", `SELECT COALESCE(percentile_cont(0.95) WITHIN GROUP
  (ORDER BY EXTRACT(EPOCH FROM now() - submitted_at)), 0)::float8` + openSuggestions},
	{"suggest.queue.oldest_seconds", `SELECT COALESCE(EXTRACT(EPOCH FROM now() - min(submitted_at)), 0)::float8` +
		openSuggestions},
}

// RegisterCommunityGauges registers the SDD 12 §6 moderation queue gauges (SDD 04 NFR-04-2, alert A9).
func RegisterCommunityGauges(m metric.Meter, query ScalarQuery) error {
	gauges := make([]metric.Float64ObservableGauge, len(communityGauges))
	observables := make([]metric.Observable, len(communityGauges))

	for i, g := range communityGauges {
		gauge, err := m.Float64ObservableGauge(g.name, metric.WithUnit(unitOf(g.name)))
		if err != nil {
			return err
		}

		gauges[i], observables[i] = gauge, gauge
	}

	_, err := m.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		for i, g := range communityGauges {
			observe(ctx, o, query, gauges[i], g.sql)
		}

		return nil
	}, observables...)

	return err
}

func unitOf(name string) string {
	if name == "suggest.queue.size" {
		return "{suggestion}"
	}

	return "s"
}

// RecordSuggestionDecision counts suggest.decisions and, past the 24 h budget, suggest.sla.breaches. decision is
// approved or rejected; age is the time from submit to the decision.
func RecordSuggestionDecision(ctx context.Context, decision string, age time.Duration) {
	instMu.Lock()
	d, b := decisions, breaches
	instMu.Unlock()

	if d == nil || decision == "" {
		return
	}

	attrs := metric.WithAttributes(attribute.String("decision", decision))
	d.Add(ctx, 1, attrs)

	if age > slaTarget {
		b.Add(ctx, 1, attrs)
	}
}
