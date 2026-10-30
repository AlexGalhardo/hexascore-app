package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"
)

const (
	sqlPGConnections = `SELECT count(*)::float8 FROM pg_stat_activity WHERE datname = current_database()`
	sqlPGSize        = `SELECT pg_database_size(current_database())::float8`
	sqlRatings24h    = `SELECT count(*)::float8 FROM ratings WHERE created_at > now() - interval '24 hours'`
	sqlLiveKeys      = `SELECT count(*)::float8 FROM api_keys WHERE revoked_at IS NULL`
)

// ScalarQuery reads one numeric SQL value. A database that cannot run the
// statement returns an error and that gauge is skipped for the scrape.
type ScalarQuery func(ctx context.Context, query string) (float64, error)

// RegisterSQLGauges registers the SQL-backed gauges from SDD 12 §6.2.
// The OTel reader invokes the callback on its own interval (60s / 5 min in production).
func RegisterSQLGauges(m metric.Meter, query ScalarQuery) error {
	conns, err := m.Float64ObservableGauge("hexa.pg.connections")
	if err != nil {
		return err
	}

	size, err := m.Float64ObservableGauge("hexa.pg.database.size_bytes")
	if err != nil {
		return err
	}

	ratings, err := m.Float64ObservableGauge("biz.ratings_24h")
	if err != nil {
		return err
	}

	keys, err := m.Float64ObservableGauge("biz.api_keys_active")
	if err != nil {
		return err
	}

	_, err = m.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		observe(ctx, o, query, conns, sqlPGConnections)
		observe(ctx, o, query, size, sqlPGSize)
		observe(ctx, o, query, ratings, sqlRatings24h)
		observe(ctx, o, query, keys, sqlLiveKeys)

		return nil
	}, conns, size, ratings, keys)

	return err
}

func observe(ctx context.Context, o metric.Observer, query ScalarQuery, g metric.Float64ObservableGauge, sql string) {
	v, err := query(ctx, sql)
	if err != nil {
		return
	}

	o.ObserveFloat64(g, v)
}
