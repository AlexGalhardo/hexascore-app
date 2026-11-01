package jobs

import (
	"cmp"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"iter"
	"regexp"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/drivers/s3"
	"github.com/AlexGalhardo/vukem/export"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/storage"
)

// ExportIDs writes the daily gzip NDJSON ID export per item type and prunes files older than 90 days;
// scheduled at 07:00 UTC in routes/schedule.go (SDD 05 FR-05-10, §7.3).
type ExportIDs struct{}

// Kind names the job on the queue.
func (ExportIDs) Kind() string { return "exports.ids" }

const exportRetentionDays = 90

// exportTypes maps the items.type value to the plural used in the object key.
var exportTypes = [...][2]string{{"game", "games"}, {"movie", "movies"}, {"series", "series"}, {"book", "books"}}

var exportKey = regexp.MustCompile(`^exports/[a-z]+_ids_(\d{4}_\d{2}_\d{2})\.json\.gz$`)

type idLine struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Wikidata  string    `json:"wikidata"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RunExportIDs writes one object per type for the UTC day of now, then deletes the objects of days older than
// exportRetentionDays. Rows stream through gzip into the disk, so no type is ever held in memory.
// popularity_prior is left out of the files until the owner confirms it as the exported popularity (A-07, Q-A11).
func RunExportIDs(ctx context.Context, db *sql.DB, disk storage.Disk, now time.Time) error {
	now = now.UTC()
	day := now.Format("2006_01_02")

	for _, t := range exportTypes {
		if err := exportTypeIDs(ctx, db, disk, t[0], fmt.Sprintf("exports/%s_ids_%s.json.gz", t[1], day)); err != nil {
			return fmt.Errorf("export ids %s: %w", t[0], err)
		}
	}

	return pruneExports(ctx, disk, now)
}

func exportTypeIDs(ctx context.Context, db *sql.DB, disk storage.Disk, itemType, key string) error {
	pr, pw := io.Pipe()
	done := make(chan error, 1)

	go func() {
		err := disk.Put(ctx, key, pr)
		_ = pr.CloseWithError(err) // unblocks the writer if Put returns early
		done <- err
	}()

	err := writeIDLines(ctx, db, pw, itemType)
	_ = pw.CloseWithError(err)

	if perr := <-done; perr != nil {
		return cmp.Or(err, perr)
	}

	return err
}

func writeIDLines(ctx context.Context, db *sql.DB, w io.Writer, itemType string) error {
	rows, err := db.QueryContext(ctx, `
SELECT id::text, slug, wikidata_qid, updated_at
FROM items WHERE type = $1 AND visibility = 'public' ORDER BY id`, itemType)
	if err != nil {
		return err
	}
	defer rows.Close()

	zw := gzip.NewWriter(w)
	if err := export.EncodeNDJSON(zw, idLines(rows)); err != nil {
		return err
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return zw.Close()
}

func idLines(rows *sql.Rows) iter.Seq2[idLine, error] {
	return func(yield func(idLine, error) bool) {
		for rows.Next() {
			var l idLine
			if err := rows.Scan(&l.ID, &l.Slug, &l.Wikidata, &l.UpdatedAt); err != nil {
				yield(l, err)

				return
			}

			l.UpdatedAt = l.UpdatedAt.UTC()

			if !yield(l, nil) {
				return
			}
		}
	}
}

func pruneExports(ctx context.Context, disk storage.Disk, now time.Time) error {
	keys, err := disk.List(ctx, "exports")
	if err != nil {
		return err
	}

	cutoff := now.Truncate(24*time.Hour).AddDate(0, 0, -exportRetentionDays)

	for _, k := range keys {
		m := exportKey.FindStringSubmatch(k)
		if m == nil {
			continue
		}

		d, err := time.Parse("2006_01_02", m[1])
		if err != nil || !d.Before(cutoff) {
			continue
		}

		if err := disk.Delete(ctx, k); err != nil {
			return err
		}
	}

	return nil
}

// idsDisk is the S3 bucket when S3_BUCKET is set (production), else a local directory for development.
func idsDisk(app *vukem.App) (storage.Disk, error) {
	if app.Source.Get("S3_BUCKET") == "" {
		d, err := storage.NewLocal(cmp.Or(app.Source.Get("ID_EXPORTS_DIR"), "storage/id-exports"), "")
		if err != nil {
			return nil, err
		}

		app.OnShutdown(func(context.Context) error { return d.Close() })

		return d, nil
	}

	cfg, err := config.Load[s3.Config](app.Source)
	if err != nil {
		return nil, err
	}

	return s3.New(cfg)
}

func handleExportIDs(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ ExportIDs) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		disk, err := idsDisk(app)
		if err != nil {
			return err
		}

		return RunExportIDs(ctx, db.SQL(), disk, app.Now())
	}, queue.Retries(2), queue.Timeout(10*time.Minute))
}
