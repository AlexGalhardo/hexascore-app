package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/dq"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// IncrementalWindow is how far back an incremental run looks when nothing says otherwise (SDD 06 §7.6).
const IncrementalWindow = 36 * time.Hour

// Stage is one job of a chain; its run is created when the previous stage succeeds.
type Stage struct {
	Job    string
	Source string
	Mode   string
}

// Chain is an ordered list of stages (SDD 06 §7.6). A stage that ends succeeded or partial (bad targets sit in
// ingest_failures) dispatches the next; failed, cancelled or skipped stops the chain.
type Chain struct {
	Name   string
	Stages []Stage
}

func fetch(source, mode string) Stage {
	return Stage{Job: "fetch." + source, Source: source, Mode: mode}
}

// The chains of the open sources. Covers join them when their job exists (P6.B.2.2); the asserts stage is last.
var (
	// Bootstrap is run by hand: Wikidata top-N, referenced entities, then the normaliser.
	Bootstrap = Chain{"bootstrap", []Stage{
		fetch("wikidata", run.ModeBootstrap),
		{Job: "fetch.entities", Source: "wikidata:entity", Mode: run.ModeBootstrap},
		{Job: "normalize", Mode: run.ModeBootstrap},
		{Job: dq.Job, Mode: run.ModeBootstrap},
	}}
	WikidataIncremental = Chain{"wikidata-incremental", []Stage{
		fetch("wikidata", run.ModeIncremental),
		{Job: "fetch.entities", Source: "wikidata:entity", Mode: run.ModeIncremental},
		{Job: "normalize", Mode: run.ModeIncremental},
		{Job: dq.Job, Mode: run.ModeIncremental},
	}}
	WikidataWeekly = Chain{"wikidata-weekly", []Stage{
		fetch("wikidata", run.ModeFull),
		{Job: "fetch.entities", Source: "wikidata:entity", Mode: run.ModeFull},
		{Job: "normalize", Mode: run.ModeFull},
		{Job: dq.Job, Mode: run.ModeFull},
	}}
	WikipediaWeekly = Chain{"wikipedia-weekly", []Stage{
		fetch("wikipedia:en", run.ModeFull),
		fetch("wikipedia:pt", run.ModeFull),
		fetch("wikipedia:es", run.ModeFull),
		{Job: "normalize", Mode: run.ModeFull},
		{Job: dq.Job, Mode: run.ModeFull},
	}}
	OpenLibraryWeekly = Chain{"openlibrary-weekly", []Stage{
		fetch("openlibrary", run.ModeFull),
		{Job: "normalize", Mode: run.ModeFull},
		{Job: dq.Job, Mode: run.ModeFull},
	}}

	TMDBWeekly = Chain{"tmdb-weekly", []Stage{
		fetch("tmdb", run.ModeFull),
		{Job: "normalize", Mode: run.ModeFull},
		{Job: dq.Job, Mode: run.ModeFull},
	}}
	// TMDBExpire refreshes records past 150 days and deletes the ones past 180, then re-normalises their items.
	TMDBExpire = Chain{"tmdb-expire", []Stage{
		{Job: "tmdb.expire", Source: "tmdb", Mode: run.ModeFull},
		{Job: "normalize", Mode: run.ModeFull},
		{Job: dq.Job, Mode: run.ModeFull},
	}}

	// EditionsBackfill walks the books once and saves the editions their stored Wikidata records list (SDD 06 §7.8).
	EditionsBackfill = Chain{"editions-backfill", []Stage{
		{Job: "backfill.editions", Mode: run.ModeBackfill},
	}}

	// TrailersWeekly re-checks every trailer candidate, then re-normalises so the verdicts reach the items. It is its
	// own chain because a stage skipped by its kill switch (ingest_youtube) stops a chain and must not stop the others.
	TrailersWeekly = Chain{"trailers-weekly", []Stage{
		fetch("youtube", run.ModeFull),
		{Job: "normalize", Mode: run.ModeFull},
		{Job: dq.Job, Mode: run.ModeFull},
	}}

	// DQDaily runs the data-quality asserts on their own (SDD 06 §7.7); every chain above ends with the same stage.
	DQDaily = Chain{"dq-daily", []Stage{{Job: dq.Job, Mode: run.ModeIncremental}}}
	// FrontendFields renders a sample of item pages and reports the missing required fields (SDD 06 §7.7).
	FrontendFields = Chain{"frontend-fields", []Stage{{Job: dq.FrontendJob, Mode: run.ModeIncremental}}}

	// Chains lists every chain by name, which is how a run's params find the next stage.
	Chains = []Chain{
		Bootstrap,
		WikidataIncremental,
		WikidataWeekly,
		WikipediaWeekly,
		OpenLibraryWeekly,
		TMDBWeekly,
		TMDBExpire,
		EditionsBackfill,
		TrailersWeekly,
		DQDaily,
		FrontendFields,
	}
)

// chainParams is the "chain" key of job_runs.params.
type chainParams struct {
	Name  string `json:"name"`
	Stage int    `json:"stage"`
}

func (c Chain) spec(stage int, params map[string]json.RawMessage, by string) (run.Spec, error) {
	p := map[string]json.RawMessage{}
	maps.Copy(p, params)

	marker, err := json.Marshal(chainParams{Name: c.Name, Stage: stage})
	if err != nil {
		return run.Spec{}, err
	}

	p["chain"] = marker

	// Fixed when the chain starts and copied to every stage, so a worker restart or a resume never moves the cursor.
	if _, ok := p["since"]; !ok && c.Stages[stage].Mode == run.ModeIncremental {
		since, err := json.Marshal(time.Now().Add(-IncrementalWindow).UTC())
		if err != nil {
			return run.Spec{}, err
		}

		p["since"] = since
	}

	raw, err := json.Marshal(p)
	if err != nil {
		return run.Spec{}, err
	}

	s := c.Stages[stage]

	return run.Spec{Job: s.Job, Source: s.Source, Mode: s.Mode, Params: raw, TriggeredBy: by}, nil
}

// Start queues the first stage. params (limit, since, types) are copied to every stage. It returns
// run.ErrRunActive when that stage is already queued or running.
func (c Chain) Start(
	ctx context.Context, db *database.DB, d run.Dispatcher, by string, params map[string]json.RawMessage,
) (run.Run, error) {
	spec, err := c.spec(0, params, by)
	if err != nil {
		return run.Run{}, err
	}

	return run.Create(ctx, db, d, spec)
}

// Advance queues the stage after run id when id belongs to a chain and ended succeeded or partial. It is safe to
// call twice for the same run: a successor already created for it is not duplicated. A successor that is already
// active (the normaliser another chain started) is enough, since its pending index covers this stage's work.
func Advance(ctx context.Context, db *database.DB, d run.Dispatcher, id int64) error {
	r, err := run.Load(ctx, db, id)
	if err != nil {
		return err
	}

	if r.Status != run.StatusSucceeded && r.Status != run.StatusPartial {
		return nil
	}

	var params map[string]json.RawMessage
	if err := json.Unmarshal(r.Params, &params); err != nil {
		return fmt.Errorf("ingest chain: run %d params: %w", id, err)
	}

	raw, ok := params["chain"]
	if !ok {
		return nil
	}

	var cp chainParams
	if err := json.Unmarshal(raw, &cp); err != nil {
		return fmt.Errorf("ingest chain: run %d chain marker: %w", id, err)
	}

	var chain Chain

	for _, c := range Chains {
		if c.Name == cp.Name {
			chain = c
		}
	}

	if chain.Name == "" || cp.Stage+1 >= len(chain.Stages) {
		return nil
	}

	by := fmt.Sprintf("chain:%d", id)

	var n int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM job_runs WHERE triggered_by = $1`, by).Scan(&n)
	if err != nil {
		return err
	}

	if n > 0 {
		return nil
	}

	spec, err := chain.spec(cp.Stage+1, params, by)
	if err != nil {
		return err
	}

	if _, err := run.Create(ctx, db, d, spec); err != nil && !errors.Is(err, run.ErrRunActive) {
		return err
	}

	return nil
}
