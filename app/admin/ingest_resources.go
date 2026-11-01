package admin

import (
	vadmin "github.com/AlexGalhardo/vukem/admin"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// JobRunsResource lists ingestion runs; triggering and cancelling go through the /admin/ingest routes (VF-008).
func JobRunsResource() *vadmin.Resource[models.JobRun] {
	return vadmin.NewResource[models.JobRun](noInputSchema).
		Slug("job-runs").
		Search("job", "source").
		Filter("status", "queued", "running", "succeeded", "partial", "failed", "skipped", "cancelled").
		List("id", "job", "source", "mode", "status", "seen", "changed", "failed", "triggered_by", "started_at", "finished_at")
}

// IngestFailuresResource lists the dead-letter targets; the key is composite so only the list is exposed.
func IngestFailuresResource() *vadmin.Resource[models.IngestFailure] {
	return vadmin.NewResource[models.IngestFailure](noInputSchema).
		Slug("ingest-failures").
		Search("external_id", "error").
		Filter("error_class", "http_4xx", "http_5xx", "timeout", "invalid_payload", "conflict").
		List("source", "external_id", "error_class", "attempts", "parked", "last_run_id", "last_seen")
}

// DQResultsResource lists data-quality assert results; list-only like the failures.
func DQResultsResource() *vadmin.Resource[models.DQResult] {
	return vadmin.NewResource[models.DQResult](noInputSchema).
		Slug("dq-results").
		Search("code").
		Filter("severity", "error", "warn").
		List("run_id", "code", "severity", "passed", "observed", "threshold", "checked_at")
}
