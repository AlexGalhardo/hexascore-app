package bootstrap

import (
	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/alerts"
)

// alertsProvider attaches the job-failure alert (A6, SDD 12 §7.5) to the queue before it is first used.
type alertsProvider struct{}

// Register adds the A6 observer.
func (alertsProvider) Register(app *vukem.App) error {
	return app.Instrument(vukem.Instrumentation{Jobs: &alerts.JobFailures{}})
}
