package jobs

import (
	"context"
	"sync"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/offers"
)

// VerifyOffers checks that the offer links still answer (RFC-003 section 3). One execution probes a chunk of
// offers and queues the next chunk with the same Since until the daily limit is reached or no offer is due; a restart
// or a timeout loses at most one chunk. It runs on the ingest queue (the worker) behind FLAG_OFFERS_VERIFY and is
// scheduled daily at 02:45.
type VerifyOffers struct {
	// Since is the start of the run: an offer checked at or after it is not due again. Zero starts a run now.
	Since time.Time `json:"since"`
	// Done is how many offers earlier chunks of this run checked.
	Done int `json:"done"`
}

// Kind names the job on the queue.
func (VerifyOffers) Kind() string { return "offers.verify" }

func handleVerifyOffers(app *vukem.App) queue.Registration {
	client := sync.OnceValue(func() *httpx.Client {
		contact := app.Source.Get("INGEST_CONTACT_URL")
		if contact == "" {
			contact = defaultContactURL
		}

		return httpx.New(httpx.Config{ContactURL: contact, Hosts: offers.HostLimits()})
	})

	return queue.Handle(func(ctx context.Context, j VerifyOffers) error {
		ff, err := app.Flags()
		if err != nil {
			return err
		}

		if !ff.Enabled(ctx, offers.VerifyFlag) {
			return nil
		}

		db, err := app.DB()
		if err != nil {
			return err
		}

		now := app.Now().UTC()

		since := j.Since
		if since.IsZero() {
			since = now
		}

		limit := min(offers.VerifyChunk, offers.VerifyDailyLimit-j.Done)
		if limit <= 0 {
			return nil
		}

		st, err := offers.Verify(ctx, db, offers.HTTPProber{Client: client().JSON}, since, now, limit)
		app.Log.InfoContext(ctx, "offers verify", "checked", st.Checked, "ok", st.OK, "redirected", st.Redirected,
			"broken", st.Broken, "gone", st.Gone, "transient", st.Transient, "done", j.Done+st.Checked)

		if err != nil {
			return err
		}

		if st.Checked < limit || j.Done+st.Checked >= offers.VerifyDailyLimit {
			return nil
		}

		q, err := app.Queue()
		if err != nil {
			return err
		}

		return q.Dispatch(ctx, VerifyOffers{Since: since, Done: j.Done + st.Checked}, queue.Delay(time.Second))
	}, queue.OnQueue("ingest"), queue.Timeout(20*time.Minute), queue.Retries(1))
}

// PruneOffers applies the offers retention (offers.Prune); scheduled daily at 03:05.
type PruneOffers struct{}

// Kind names the job on the queue.
func (PruneOffers) Kind() string { return "offers.prune" }

func handlePruneOffers(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ PruneOffers) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		p, err := offers.Prune(ctx, db, app.Now().UTC())
		if err != nil {
			return err
		}

		app.Log.InfoContext(ctx, "prune offers", "deactivated", p.Deactivated, "deleted", p.Deleted, "clicks", p.Clicks)

		return nil
	}, queue.Timeout(10*time.Minute), queue.Retries(1))
}
