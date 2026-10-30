package metrics

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestSQLGauges(t *testing.T) {
	t.Parallel()

	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))
	query := func(_ context.Context, sql string) (float64, error) {
		switch {
		case strings.Contains(sql, "pg_stat_activity"):
			return 4, nil
		case strings.Contains(sql, "pg_database_size"):
			return 4096, nil
		case strings.Contains(sql, "ratings"):
			return 7, nil
		case strings.Contains(sql, "api_keys"):
			return 2, nil
		default:
			t.Errorf("unexpected sql %s", sql)

			return 0, nil
		}
	}

	if err := RegisterSQLGauges(mp.Meter(meterName), query); err != nil {
		t.Fatal(err)
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}

	got := map[string]float64{}

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			gauge, ok := m.Data.(metricdata.Gauge[float64])
			if !ok || len(gauge.DataPoints) == 0 {
				continue
			}

			got[m.Name] = gauge.DataPoints[0].Value
		}
	}

	want := map[string]float64{
		"hexa.pg.connections":         4,
		"hexa.pg.database.size_bytes": 4096,
		"biz.ratings_24h":             7,
		"biz.api_keys_active":         2,
	}

	for name, value := range want {
		if got[name] != value {
			t.Errorf("%s = %v, want %v", name, got[name], value)
		}
	}
}
