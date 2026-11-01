package tmdb

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/expire"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// JobName is the job_runs.job of the expiry run.
const JobName = "tmdb.expire"

// RefreshAfter and DeleteAfter are the TMDB cache terms (SDD 06 §7.6, FR-06-5): no value may be older than 180
// days by checked_at, so refresh starts 30 days earlier.
const (
	RefreshAfter = 150 * 24 * time.Hour
	DeleteAfter  = 180 * 24 * time.Hour
)

// ExpireStep deletes the TMDB records past DeleteAfter with their TMDB covers and refreshes the ones past
// RefreshAfter through c. The item supplies the type; the stored payload supplies the TMDB id.
func ExpireStep(c *Connector, client *ingest.Client, now func() time.Time) run.Step {
	return expire.Step(expire.Policy{
		Source:       c.Source(),
		RefreshAfter: RefreshAfter,
		DeleteAfter:  DeleteAfter,
		Target: func(qid, itemType string, payload []byte) ingest.Target {
			var p struct {
				ID int `json:"id"`
			}

			_ = json.Unmarshal(payload, &p)

			id := ""
			if p.ID > 0 {
				id = strconv.Itoa(p.ID)
			}

			return ingest.Target{ID: qid, Type: Ref(itemType, id, "")}
		},
	}, c, client, now)
}
