package metrics

import (
	"context"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestLabelCap(t *testing.T) {
	t.Parallel()

	kept := NewCap(ErrorTypeCap)

	for i := range ErrorTypeCap {
		name := fmt.Sprintf("e%d", i)
		if got := kept.Observe(name); got != name {
			t.Fatalf("Observe(%s) = %s", name, got)
		}
	}

	if got := kept.Observe("e0"); got != "e0" {
		t.Fatalf("repeat = %s", got)
	}

	if got := kept.Observe("extra"); got != "other" {
		t.Fatalf("overflow = %s", got)
	}

	if got := kept.Observe("extra-2"); got != "other" {
		t.Fatalf("second overflow = %s", got)
	}

	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))
	counter, err := newErrorCounter(mp.Meter(meterName))
	if err != nil {
		t.Fatal(err)
	}

	for i := range ErrorTypeCap + 1 {
		counter.add(context.Background(), fmt.Sprintf("t%d", i))
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}

	values := map[string]bool{}

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "hexa.errors" {
				continue
			}

			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("hexa.errors data %T", m.Data)
			}

			for _, dp := range sum.DataPoints {
				for _, attr := range dp.Attributes.ToSlice() {
					if attr.Key == "error.type" {
						values[attr.Value.AsString()] = true
					}
				}
			}
		}
	}

	if !values["other"] {
		t.Fatal("hexa.errors missing other")
	}

	if values["t50"] {
		t.Fatal("51st error.type was kept")
	}

	if !values["t0"] || !values["t49"] {
		t.Fatalf("kept set = %v", values)
	}
}

func TestRouteGroup(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"/":                   "web.public",
		"/games":              "web.public",
		"/games/{slug}":       "web.public",
		"/dashboard":          "web.app",
		"/profile/api":        "web.app",
		"/admin/ingestion":    "admin",
		"/api/v1/{type}":      "api.v1",
		"/v1/games":           "api.v1",
		"/app/v1/tokens":      "app.v1",
		"/search":             "search",
		"/api/search/suggest": "search",
		"/up":                 "health",
		"GET /up":             "health",
	}

	for route, want := range cases {
		if got := RouteGroup(route); got != want {
			t.Errorf("RouteGroup(%s) = %s, want %s", route, got, want)
		}
	}
}
