package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/vukemtest"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/AlexGalhardo/hexascore-app/app/metrics"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// SDD 12 §6: the moderation gauges come from SQL, the decision counters from the approve and reject paths.
func TestMetrics_CommunityGauges(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))
	vukemtest.New(t, bootstrap.New, vukemtest.Options{
		AppOptions: []vukem.Option{
			vukem.WithRoutes(func(_ *router.Router, built *vukem.App) error {
				metrics.Bind(built, mp)

				return nil
			}),
		},
	})

	query := func(_ context.Context, sql string) (float64, error) {
		switch {
		case strings.Contains(sql, "percentile_cont(0.5)"):
			return 3600, nil
		case strings.Contains(sql, "percentile_cont(0.95)"):
			return 7200, nil
		case strings.Contains(sql, "min(submitted_at)"):
			return 9000, nil
		default:
			return 4, nil
		}
	}
	if err := metrics.RegisterCommunityGauges(mp.Meter("hexascore"), query); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	metrics.RecordSuggestionDecision(ctx, "approved", time.Hour)
	metrics.RecordSuggestionDecision(ctx, "rejected", 25*time.Hour)
	metrics.RecordSuggestionDecision(ctx, "rejected", 2*time.Hour)

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatal(err)
	}

	gauges := map[string]float64{}
	counters := map[string]map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch d := m.Data.(type) {
			case metricdata.Gauge[float64]:
				if len(d.DataPoints) > 0 {
					gauges[m.Name] = d.DataPoints[0].Value
				}
			case metricdata.Sum[int64]:
				counters[m.Name] = map[string]int64{}
				for _, dp := range d.DataPoints {
					v, _ := dp.Attributes.Value("decision")
					counters[m.Name][v.AsString()] = dp.Value
				}
			}
		}
	}

	for name, want := range map[string]float64{
		"suggest.queue.size": 4, "suggest.queue.age_p50_seconds": 3600, "suggest.queue.age_p95_seconds": 7200,
		"suggest.queue.oldest_seconds": 9000,
	} {
		if gauges[name] != want {
			t.Errorf("%s = %v, want %v (all %v)", name, gauges[name], want, gauges)
		}
	}

	if d := counters["suggest.decisions"]; d["approved"] != 1 || d["rejected"] != 2 {
		t.Errorf("suggest.decisions = %v", d)
	}
	if b := counters["suggest.sla.breaches"]; b["rejected"] != 1 || b["approved"] != 0 {
		t.Errorf("suggest.sla.breaches = %v", b)
	}
}
