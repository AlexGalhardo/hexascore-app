package jobs

import (
	"context"
	"fmt"
	"os"
	"strings"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/eval"
)

// RecoEvalMonthly checks the collaborative-filtering trigger (SDD 08 section 7.6); scheduled on day 1 at 05:30 in
// routes/schedule.go. The CSV goes to RECO_EVAL_DIR when set, else only to the log.
type RecoEvalMonthly struct{}

// Kind names the job on the queue.
func (RecoEvalMonthly) Kind() string { return "recommend.eval_monthly" }

func handleRecoEvalMonthly(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ RecoEvalMonthly) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		now := app.Now().UTC()
		w := new(strings.Builder)

		t, err := eval.RunMonthly(ctx, db, now, eval.Thresholds{}, w)
		if err != nil {
			return err
		}

		app.Log.InfoContext(ctx, "reco eval monthly", "users", t.Users, "triggered", t.Triggered)

		dir := app.Source.Get("RECO_EVAL_DIR")
		if dir == "" {
			return nil
		}

		root, err := os.OpenRoot(dir)
		if err != nil {
			return fmt.Errorf("open reco eval dir: %w", err)
		}

		defer func() { _ = root.Close() }()

		return root.WriteFile("reco-eval-"+now.Format("2006-01")+".csv", []byte(w.String()), 0o600)
	})
}
