package covers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

const (
	// JobName is the job_runs.job of the covers job; it is also its kill switch (ingest_covers) and its
	// ingest_failures source.
	JobName = "covers"
	// BatchSize is how many items one step looks at; downloads and encodes run inside the step's transaction.
	BatchSize = 25
	// Workers is how many items download at once; cwebp itself runs one at a time (NFR-06-2).
	Workers = 8
	// incrementalWindow matches ingest.IncrementalWindow (not imported: ingest imports this package).
	incrementalWindow = 36 * time.Hour
)

// ErrNoDisk means the job has no bucket to write to (S3_BUCKET unset).
var ErrNoDisk = errors.New("covers: no storage disk (S3_BUCKET is not set)")

// Job is the covers job (SDD 06 §7.5, port of covers.py): it picks the best cover the records name, stores the
// two WebP variants under content-hashed keys and points the item at them.
type Job struct {
	Disk storage.Disk
	// HTTP downloads images and calls Commons: the ingest client's Image client (httpclient.Safe, 15 MB cap).
	HTTP    *http.Client
	Encoder Encoder
	// CommonsAPI replaces the Commons endpoint (tests).
	CommonsAPI string
	Now        func() time.Time
	// Workers and Batch override the package defaults (tests).
	Workers, Batch int
	// ItemsChanged is told which items got a new cover, in the step's transaction; nil skips it.
	ItemsChanged func(ctx context.Context, ex database.Executor, ids []string) error
	// Reserve records a key in cover_gc outside the step's transaction before its object is written, so an
	// object whose transaction rolls back is still reaped by GC (which forgets reserved keys an item points at).
	Reserve func(ctx context.Context, key string) error
}

// Reserver is the Reserve func that writes cover_gc through db, a handle outside the step's transaction.
func Reserver(db database.Executor) func(ctx context.Context, key string) error {
	return func(ctx context.Context, key string) error {
		_, err := db.ExecContext(ctx, `INSERT INTO cover_gc (key) VALUES ($1) ON CONFLICT (key) DO NOTHING`, key)

		return err
	}
}

// Target is one item and its stored cover columns.
type Target struct {
	ItemID, QID, Type string
	Records           []store.Record
	// Key, ThumbKey and Source are the current cover_key, cover_thumb_key and cover_source ("" when null).
	Key, ThumbKey, Source string
}

// Patch is the cover an item gets.
type Patch struct {
	ItemID, Key, ThumbKey, Source string
	Attribution                   map[string]string
}

// Outcome is what Run did for one target: a nil Patch and Err mean nothing changes.
type Outcome struct {
	Patch *Patch
	Err   error
}

// Key is the immutable object key of a variant: covers/<id>-<sha256(source)[:16]><suffix>.webp.
func Key(id string, source []byte, suffix string) string {
	sum := sha256.Sum256(source)

	return "covers/" + id + "-" + hex.EncodeToString(sum[:8]) + suffix + ".webp"
}

// Store downloads src and writes each variant whose current key does not already hold this source's hash, so a
// cover stored before thumbnails existed only gets its thumbnail.
func (j *Job) Store(ctx context.Context, itemID, src, curKey, curThumb string) (key, thumb string, err error) {
	body, err := j.get(ctx, src)
	if err != nil {
		return "", "", err
	}

	out := [2]string{}

	for i, v := range []struct {
		cur, suffix string
		variant     Variant
	}{{curKey, "", Full}, {curThumb, "-t", Thumb}} {
		k := Key(itemID, body, v.suffix)
		out[i] = k

		if v.cur == k {
			continue
		}

		webp, err := j.Encoder.Encode(ctx, body, v.variant)
		if err != nil {
			return "", "", fmt.Errorf("encode %s: %w", k, err)
		}

		if j.Reserve != nil {
			if err := j.Reserve(ctx, k); err != nil {
				return "", "", fmt.Errorf("reserve %s: %w", k, err)
			}
		}

		if err := storage.PutBytes(ctx, j.Disk, k, webp); err != nil {
			return "", "", fmt.Errorf("put %s: %w", k, err)
		}
	}

	return out[0], out[1], nil
}

// Cover decides and stores the cover of one target. Source order: Open Library (books), TMDB/IGDB, Steam, GOG,
// then the Commons P3383 poster; a cover from the same source with both keys set is kept without a download.
func (j *Job) Cover(ctx context.Context, t Target) (*Patch, error) {
	both := t.Key != "" && t.ThumbKey != ""

	c, err := normalize.PickCover(t.QID, t.Type, t.Records)
	if err != nil {
		return nil, err
	}

	if c == nil {
		file := CommonsFile(t.QID, t.Records)
		if file == "" || (both && t.Source == normalize.SourceCommons) {
			return nil, nil //nolint:nilnil // nothing to do is not an error
		}

		if c, err = j.commonsCover(ctx, file); err != nil || c == nil {
			return nil, err
		}
	}

	if both && c.Source == t.Source {
		return nil, nil //nolint:nilnil // current cover
	}

	key, thumb, err := j.Store(ctx, t.ItemID, c.URL, t.Key, t.ThumbKey)
	if err != nil {
		return nil, err
	}

	if key == t.Key && thumb == t.ThumbKey && c.Source == t.Source {
		return nil, nil //nolint:nilnil // same bytes
	}

	return &Patch{ItemID: t.ItemID, Key: key, ThumbKey: thumb, Source: c.Source, Attribution: c.Attribution}, nil
}

// Run covers the targets concurrently; one target's failure never touches the others.
func (j *Job) Run(ctx context.Context, targets []Target) []Outcome {
	out := make([]Outcome, len(targets))
	n := j.Workers

	if n <= 0 {
		n = Workers
	}

	slots := make(chan struct{}, n)

	var wg sync.WaitGroup

	for i, t := range targets {
		wg.Go(func() {
			slots <- struct{}{}
			defer func() { <-slots }()

			p, err := j.Cover(ctx, t)
			out[i] = Outcome{Patch: p, Err: err}
		})
	}

	wg.Wait()

	return out
}

type params struct {
	Since time.Time `json:"since"`
	QIDs  []string  `json:"qids"`
}

type cursor struct {
	After string `json:"after"`
}

// Step covers one keyset page of items. An incremental run looks only at items without both keys or with a
// record fetched since the window; a full run looks at every item and, when it was not restricted to qids,
// finally drops the Commons covers that no P3383 poster backs (they came from P18 in the legacy app).
func (j *Job) Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	if j.Disk == nil {
		return run.Result{}, fmt.Errorf("%w: %w", ErrNoDisk, queue.ErrPermanent)
	}

	var p params
	if len(r.Params) > 0 {
		if err := json.Unmarshal(r.Params, &p); err != nil {
			return run.Result{}, fmt.Errorf("covers: params: %w: %w", err, queue.ErrPermanent)
		}
	}

	var cur cursor
	if len(r.Cursor) > 0 {
		if err := json.Unmarshal(r.Cursor, &cur); err != nil {
			return run.Result{}, fmt.Errorf("covers: cursor: %w: %w", err, queue.ErrPermanent)
		}
	}

	var since *time.Time

	if r.Mode == run.ModeIncremental {
		if p.Since.IsZero() {
			p.Since = j.now().Add(-incrementalWindow)
		}

		since = &p.Since
	}

	batch := j.Batch
	if batch <= 0 {
		batch = BatchSize
	}

	targets, err := loadTargets(ctx, tx, cur.After, since, p.QIDs, batch)
	if err != nil {
		return run.Result{}, err
	}

	res := run.Result{Done: len(targets) < batch, Targets: len(targets), Seen: len(targets)}

	changed, err := j.apply(ctx, tx, r, j.Run(ctx, targets), targets, &res)
	if err != nil {
		return run.Result{}, err
	}

	if res.Done && r.Mode == run.ModeFull && len(p.QIDs) == 0 {
		dropped, err := dropLooseCommons(ctx, tx)
		if err != nil {
			return run.Result{}, err
		}

		changed = append(changed, dropped...)
	}

	if len(changed) > 0 && j.ItemsChanged != nil {
		if err := j.ItemsChanged(ctx, tx, changed); err != nil {
			return run.Result{}, err
		}
	}

	if len(targets) > 0 {
		cur.After = targets[len(targets)-1].ItemID
	}

	res.Cursor, err = json.Marshal(cur)

	return res, err
}

func (j *Job) now() time.Time {
	if j.Now != nil {
		return j.Now()
	}

	return time.Now()
}

// apply writes the outcomes: patches update the item, failures go to the dead letter keyed by QID.
func (j *Job) apply(
	ctx context.Context, tx *database.Tx, r run.Run, outs []Outcome, targets []Target, res *run.Result,
) ([]string, error) {
	var changed []string

	for i, o := range outs {
		t := targets[i]

		if o.Err != nil {
			// A missing cwebp fails every item alike: stop the run instead of parking the whole catalogue.
			if errors.Is(o.Err, ErrNoEncoder) {
				return nil, o.Err
			}

			if _, err := run.RecordFailure(ctx, tx, r.ID, JobName, t.QID, "cover", o.Err); err != nil {
				return nil, err
			}

			res.Failed++

			continue
		}

		if err := run.ClearFailure(ctx, tx, JobName, t.QID); err != nil {
			return nil, err
		}

		if o.Patch == nil {
			continue
		}

		var att any
		if o.Patch.Attribution != nil {
			att = o.Patch.Attribution
		}

		raw, err := json.Marshal(att)
		if err != nil {
			return nil, err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE items SET cover_key = $2, cover_thumb_key = $3, cover_source = $4,
cover_attribution = $5::jsonb WHERE id = $1`, t.ItemID, o.Patch.Key, o.Patch.ThumbKey, o.Patch.Source, string(raw)); err != nil {
			return nil, fmt.Errorf("covers: update %s: %w", t.ItemID, err)
		}

		res.Changed++

		changed = append(changed, t.ItemID)
	}

	return changed, nil
}

// communityCover is true when the latest audit row of the item's cover is a community edit; those are never
// replaced by ingestion.
const communityCover = `coalesce((SELECT a.source::text FROM catalog_audit a
  WHERE a.target_type = 'item' AND a.target_id = i.id AND a.field = 'cover' ORDER BY a.id DESC LIMIT 1), '') = 'community'`

func loadTargets(
	ctx context.Context, tx *database.Tx, after string, since *time.Time, qids []string, limit int,
) ([]Target, error) {
	if after == "" {
		after = "00000000-0000-0000-0000-000000000000"
	}

	var list any
	if len(qids) > 0 {
		list = qids
	}

	rows, err := tx.QueryContext(ctx, `SELECT i.id::text, i.wikidata_qid, i.type, coalesce(i.cover_key, ''),
  coalesce(i.cover_thumb_key, ''), coalesce(i.cover_source, '')
FROM items i
WHERE i.id > $1::uuid
  AND ($2::text[] IS NULL OR i.wikidata_qid = ANY($2::text[]))
  AND ($2::text[] IS NOT NULL OR NOT EXISTS (SELECT 1 FROM ingest_failures f
    WHERE f.source = 'covers' AND f.external_id = i.wikidata_qid AND f.parked))
  AND ($3::timestamptz IS NULL OR i.cover_key IS NULL OR i.cover_thumb_key IS NULL
    OR EXISTS (SELECT 1 FROM source_records s WHERE s.external_id = i.wikidata_qid AND s.fetched_at >= $3))
  AND NOT `+communityCover+`
ORDER BY i.id LIMIT $4`, after, list, since, limit)
	if err != nil {
		return nil, fmt.Errorf("covers: list items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var (
		out  []Target
		keys []string
	)

	for rows.Next() {
		var t Target
		if err := rows.Scan(&t.ItemID, &t.QID, &t.Type, &t.Key, &t.ThumbKey, &t.Source); err != nil {
			return nil, err
		}

		out = append(out, t)
		keys = append(keys, t.QID)
	}

	if err := rows.Err(); err != nil || len(out) == 0 {
		return out, err
	}

	recs, err := loadRecords(ctx, tx, keys)
	if err != nil {
		return nil, err
	}

	for i := range out {
		out[i].Records = recs[out[i].QID]
	}

	return out, nil
}

// loadRecords reads only the sources that can name a cover.
func loadRecords(ctx context.Context, tx *database.Tx, qids []string) (map[string][]store.Record, error) {
	rows, err := tx.QueryContext(ctx, `SELECT source, external_id, payload FROM source_records
WHERE external_id = ANY($1::text[]) AND source IN ('wikidata', 'openlibrary', 'tmdb', 'igdb', 'steam', 'gog')`, qids)
	if err != nil {
		return nil, fmt.Errorf("covers: read records: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := map[string][]store.Record{}

	for rows.Next() {
		var r store.Record
		if err := rows.Scan(&r.Source, &r.ExternalID, &r.Payload); err != nil {
			return nil, err
		}

		out[r.ExternalID] = append(out[r.ExternalID], r)
	}

	return out, rows.Err()
}

// dropLooseCommons clears Commons covers of items without a P3383 poster (legacy covers.py
// drop_loose_commons_covers); the cover_gc trigger schedules their objects for deletion.
func dropLooseCommons(ctx context.Context, tx *database.Tx) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `UPDATE items i SET cover_key = NULL, cover_thumb_key = NULL, cover_source = NULL,
  cover_attribution = NULL
WHERE i.cover_source = 'commons' AND NOT `+communityCover+`
  AND NOT EXISTS (SELECT 1 FROM source_records w WHERE w.source = 'wikidata' AND w.external_id = i.wikidata_qid
    AND w.payload #>> '{claims,P3383,0,value}' IS NOT NULL)
RETURNING i.id::text`)
	if err != nil {
		return nil, fmt.Errorf("covers: drop loose commons covers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}
