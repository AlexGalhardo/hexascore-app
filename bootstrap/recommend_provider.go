package bootstrap

import (
	"context"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/events"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/primitives"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// recommendProvider feeds search clicks to the trending counter (SDD 08 open issue 08-5), queues rebuilds on bans and
// exposes the snapshot age.
type recommendProvider struct{}

// Register subscribes to search.ResultClicked; only the item id is used.
func (recommendProvider) Register(app *vukem.App) error {
	rec := primitives.For(app)
	app.Events.Register(events.On(func(_ context.Context, e search.ResultClicked) error {
		rec.CountClick(e.ItemID, e.At)

		return nil
	}))

	app.Events.Register(trending.Listeners(app)...)

	// Age of the newest snapshot, read at scrape time; alert A14 fires above 3 h. No data while none exists.
	_, err := otel.Meter("hexascore").Float64ObservableGauge("trending.snapshot_age_seconds", metric.WithUnit("s"),
		metric.WithFloat64Callback(func(ctx context.Context, o metric.Float64Observer) error {
			if db, derr := app.DB(); derr == nil {
				if age, aerr := trending.AgeSeconds(ctx, db, time.Now()); aerr == nil {
					o.Observe(age)
				}
			}

			return nil
		}))

	return err
}
