package tests

import (
	"context"
	"testing"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/vukemtest"
	"github.com/AlexGalhardo/vukem/web"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/AlexGalhardo/hexascore-app/app/metrics"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestMetrics_EveryRouteGroupEmitsLatency(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		AppOptions: []vukem.Option{
			vukem.WithRoutes(func(r *router.Router, built *vukem.App) error {
				metrics.Bind(built, mp)
				// Search and /api/v1 mount only on Postgres (VF-033). These probes
				// are the route patterns the groups are derived from.
				probe := func(*web.Ctx) error { return nil }
				r.Get("/search", probe)
				r.Get("/api/v1/games", probe)

				return nil
			}),
		},
	})
	c := app.Client(t)

	for _, path := range []string{
		"/",
		"/dashboard",
		"/admin/ingestion",
		"/api/v1/games",
		"/app/v1/tokens",
		"/search",
		"/up",
	} {
		c.Get(path)
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}

	groups := map[string]int{}

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http.server.request.duration" {
				continue
			}

			hist, ok := m.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatalf("duration data %T", m.Data)
			}

			for _, dp := range hist.DataPoints {
				route := ""

				for _, attr := range dp.Attributes.ToSlice() {
					if string(attr.Key) == "http.route" {
						route = attr.Value.AsString()
					}
				}

				groups[metrics.RouteGroup(route)]++
			}
		}
	}

	for _, name := range []string{"web.public", "web.app", "admin", "api.v1", "app.v1", "search", "health"} {
		if groups[name] != 1 {
			t.Errorf("group %s points = %d, all = %v", name, groups[name], groups)
		}
	}
}
