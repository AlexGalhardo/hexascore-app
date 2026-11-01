package igdb

import (
	"encoding/json"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/expire"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// JobName is the job_runs.job of the refresh run.
const JobName = "igdb.refresh"

// RefreshAfter and DeleteAfter are the D-36 interim terms (SDD 06 §7.1): records are fetched again after 24 h and
// nothing IGDB-sourced, cover included, is kept past 7 days (DQ-I-13).
const (
	RefreshAfter = 24 * time.Hour
	DeleteAfter  = 7 * 24 * time.Hour
)

// RefreshStep deletes the IGDB records past DeleteAfter with their IGDB covers and refreshes the ones past
// RefreshAfter through c, by the slug the stored payload carries.
func RefreshStep(c *Connector, client *ingest.Client, now func() time.Time) run.Step {
	return expire.Step(expire.Policy{
		Source:       c.Source(),
		RefreshAfter: RefreshAfter,
		DeleteAfter:  DeleteAfter,
		Target: func(qid, _ string, payload []byte) ingest.Target {
			var p struct {
				Slug string `json:"slug"`
			}

			_ = json.Unmarshal(payload, &p)

			return ingest.Target{ID: qid, Type: p.Slug}
		},
	}, c, client, now)
}
