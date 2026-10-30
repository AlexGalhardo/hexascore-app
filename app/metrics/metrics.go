package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "hexascore"

const slowQuery = 200 * time.Millisecond

// bound keeps a test's meter on the app that asked for it, so a parallel
// test does not publish into another app's reader.
var bound sync.Map

// Bind selects the meter provider Install uses for app.
func Bind(app *vukem.App, mp metric.MeterProvider) {
	if app != nil && mp != nil {
		bound.Store(app, mp)
	}
}

type errorCounter struct {
	c   metric.Int64Counter
	cap *Cap
}

func newErrorCounter(m metric.Meter) (*errorCounter, error) {
	c, err := m.Int64Counter("hexa.errors", metric.WithDescription("Application errors by type."))
	if err != nil {
		return nil, err
	}

	return &errorCounter{c: c, cap: NewCap(ErrorTypeCap)}, nil
}

func (e *errorCounter) add(ctx context.Context, errorType string) {
	if e == nil || e.c == nil {
		return
	}

	e.c.Add(ctx, 1, metric.WithAttributes(attribute.String("error.type", e.cap.Observe(errorType))))
}

// SetBackup records the last successful dump for the backup gauges.
func SetBackup(at time.Time, size int64) {
	backupAt.Store(at.UnixNano())
	backupSize.Store(size)
	backupSet.Store(true)
}

var (
	instMu     sync.Mutex
	errs       *errorCounter
	runs       metric.Int64Counter
	signups    metric.Int64Counter
	verified   metric.Int64Counter
	ssrFall    metric.Int64Counter
	slow       metric.Int64Counter
	backupSet  atomic.Bool
	backupAt   atomic.Int64
	backupSize atomic.Int64
)

// Install registers the platform instruments on app and the HTTP middleware
// that emits http.server.request.duration. drivers/otel is not in the adopted
// vukem module yet (P2.C.2.1); this middleware is that sample until then.
func Install(app *vukem.App, age func() time.Duration) error {
	mp := otel.GetMeterProvider()
	if v, ok := bound.Load(app); ok {
		boundMP, isMP := v.(metric.MeterProvider)
		if !isMP {
			return fmt.Errorf("metrics: bound provider has type %T", v)
		}

		mp = boundMP
	}

	m := mp.Meter(meterName)
	counter, err := newErrorCounter(m)
	if err != nil {
		return err
	}

	scheduleRuns, err := m.Int64Counter("hexa.schedule.runs")
	if err != nil {
		return err
	}

	signupC, err := m.Int64Counter("biz.signups")
	if err != nil {
		return err
	}

	verifiedC, err := m.Int64Counter("biz.verified")
	if err != nil {
		return err
	}

	fallbacks, err := m.Int64Counter("hexa.ssr.fallbacks")
	if err != nil {
		return err
	}

	slowC, err := m.Int64Counter("hexa.db.slow_queries")
	if err != nil {
		return err
	}

	instMu.Lock()
	errs, runs, signups, verified, ssrFall, slow = counter, scheduleRuns, signupC, verifiedC, fallbacks, slowC
	instMu.Unlock()

	if err := registerHistograms(m); err != nil {
		return err
	}

	if err := registerPool(m, app); err != nil {
		return err
	}

	if err := registerBackup(m); err != nil {
		return err
	}

	if err := registerHeartbeat(m, age); err != nil {
		return err
	}

	db, err := app.DB()
	if err != nil {
		return err
	}

	if err := RegisterSQLGauges(m, func(ctx context.Context, query string) (float64, error) {
		return scalar(ctx, db, query)
	}); err != nil {
		return err
	}

	db.AddQueryObserver(slowObserver{})

	inflight, err := m.Int64UpDownCounter("hexa.http.inflight")
	if err != nil {
		return err
	}

	duration, err := m.Float64Histogram("http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of HTTP server requests."),
	)
	if err != nil {
		return err
	}

	app.Router.UseServer(inflightMiddleware(inflight))
	app.Router.UseServer(durationMiddleware(duration))

	return nil
}

func registerHistograms(m metric.Meter) error {
	var err error
	if _, err = m.Float64Histogram("hexa.ssr.render.duration", metric.WithUnit("s")); err != nil {
		return err
	}

	return nil
}

func registerPool(m metric.Meter, app *vukem.App) error {
	db, err := app.DB()
	if err != nil {
		return err
	}

	sqlDB := db.SQL()
	inUse, err := m.Int64ObservableGauge("hexa.db.pool.in_use")
	if err != nil {
		return err
	}

	maxOpen, err := m.Int64ObservableGauge("hexa.db.pool.max")
	if err != nil {
		return err
	}

	waits, err := m.Int64ObservableGauge("hexa.db.pool.wait_count")
	if err != nil {
		return err
	}

	_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		stats := sqlDB.Stats()
		o.ObserveInt64(inUse, int64(stats.InUse))
		o.ObserveInt64(maxOpen, int64(stats.MaxOpenConnections))
		o.ObserveInt64(waits, stats.WaitCount)

		return nil
	}, inUse, maxOpen, waits)

	return err
}

func registerBackup(m metric.Meter) error {
	age, err := m.Float64ObservableGauge("backup.last_success_age_seconds", metric.WithUnit("s"))
	if err != nil {
		return err
	}

	size, err := m.Int64ObservableGauge("backup.dump_size_bytes")
	if err != nil {
		return err
	}

	// Placeholders until a dump records a success. No observation means "never".
	_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		if !backupSet.Load() {
			return nil
		}

		o.ObserveFloat64(age, time.Since(time.Unix(0, backupAt.Load())).Seconds())
		o.ObserveInt64(size, backupSize.Load())

		return nil
	}, age, size)

	return err
}

func registerHeartbeat(m metric.Meter, age func() time.Duration) error {
	if age == nil {
		return nil
	}

	g, err := m.Float64ObservableGauge("hexa.scheduler.heartbeat_age_seconds", metric.WithUnit("s"))
	if err != nil {
		return err
	}

	_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		o.ObserveFloat64(g, age().Seconds())

		return nil
	}, g)

	return err
}

func scalar(ctx context.Context, db *database.DB, query string) (float64, error) {
	var v float64
	err := db.QueryRowContext(ctx, query).Scan(&v)

	return v, err
}

func current() (e *errorCounter, schedule, signup, verifiedC, fallback metric.Int64Counter) {
	instMu.Lock()
	defer instMu.Unlock()

	return errs, runs, signups, verified, ssrFall
}

// RecordError counts hexa.errors, collapsing types past the cap to other.
func RecordError(ctx context.Context, errorType string) {
	e, _, _, _, _ := current()
	if e != nil {
		e.add(ctx, errorType)
	}
}

// RecordSchedule counts one hexa.schedule.runs sample.
func RecordSchedule(ctx context.Context, task, outcome string) {
	_, schedule, _, _, _ := current()
	if schedule == nil {
		return
	}

	schedule.Add(ctx, 1, metric.WithAttributes(
		attribute.String("task", task),
		attribute.String("outcome", outcome),
	))
}

// RecordSignup counts biz.signups.
func RecordSignup(ctx context.Context) {
	_, _, signup, _, _ := current()
	if signup != nil {
		signup.Add(ctx, 1)
	}
}

// RecordVerified counts biz.verified.
func RecordVerified(ctx context.Context) {
	_, _, _, verifiedC, _ := current()
	if verifiedC != nil {
		verifiedC.Add(ctx, 1)
	}
}

// RecordSSRFallback counts hexa.ssr.fallbacks. The hook has the page URL, not
// the mux pattern, so the route label is the path with query stripped.
func RecordSSRFallback(ctx context.Context, rawURL string, public bool) {
	_, _, _, _, fallback := current()
	if fallback == nil {
		return
	}

	path := rawURL
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}

	flag := "false"
	if public {
		flag = "true"
	}

	fallback.Add(ctx, 1, metric.WithAttributes(
		attribute.String("http.route", path),
		attribute.String("public", flag),
	))
}

type slowObserver struct{}

func (slowObserver) ObserveQuery(ctx context.Context, q database.Query) {
	instMu.Lock()
	counter := slow
	instMu.Unlock()

	if counter == nil || q.Duration <= slowQuery {
		return
	}

	fields := strings.Fields(strings.TrimSpace(q.SQL))
	if len(fields) == 0 {
		return
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", strings.ToUpper(fields[0]))))
}

func inflightMiddleware(c metric.Int64UpDownCounter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Add(r.Context(), 1)
			defer c.Add(r.Context(), -1)
			next.ServeHTTP(w, r)
		})
	}
}

func durationMiddleware(h metric.Float64Histogram) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rw, r)
			route := routeOf(r.Pattern)
			h.Record(r.Context(), time.Since(start).Seconds(), metric.WithAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", route),
				attribute.Int("http.response.status_code", rw.code()),
			))
		})
	}
}

func routeOf(pattern string) string {
	if _, rest, ok := strings.Cut(pattern, " "); ok {
		pattern = strings.TrimSpace(rest)
	}

	i := strings.IndexByte(pattern, '/')
	if i < 0 {
		return ""
	}

	return strings.TrimSuffix(pattern[i:], "{$}")
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}

	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func (s *statusRecorder) code() int {
	if s.status == 0 {
		return http.StatusOK
	}

	return s.status
}
