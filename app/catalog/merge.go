package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
)

// Merge errors; the admin route (P1.G.5.1) maps them to 409/404.
var (
	ErrMergeSelf         = errors.New("catalog: cannot merge an item into itself")
	ErrMergeNotFound     = errors.New("catalog: merge item not found")
	ErrMergeTypeMismatch = errors.New("catalog: merge items have different types")
	ErrMergeRedirected   = errors.New("catalog: merge source was already merged away")
)

// Merged is the `catalog.merged` job dispatched in the merge transaction;
// score recompute (P1.D.3.2), list and status moves (P1.F.1.2) and reindex
// (SDD 07) subscribe to it.
type Merged struct {
	Type   string `json:"type"`
	FromID string `json:"from_id"`
	IntoID string `json:"into_id"`
}

// Kind names the job on the queue.
func (Merged) Kind() string { return "catalog.merged" }

// snapshotSQL captures one item with everything a manual restore needs.
const snapshotSQL = `SELECT jsonb_build_object(
  'item', to_jsonb(i),
  'translations', (SELECT coalesce(jsonb_agg(to_jsonb(x)), '[]') FROM item_translations x WHERE x.item_id = i.id),
  'alt_titles', (SELECT coalesce(jsonb_agg(to_jsonb(x)), '[]') FROM item_alt_titles x WHERE x.item_id = i.id),
  'entities', (SELECT coalesce(jsonb_agg(to_jsonb(x)), '[]') FROM item_entities x WHERE x.item_id = i.id),
  'links', (SELECT coalesce(jsonb_agg(to_jsonb(x)), '[]') FROM item_links x
    WHERE x.from_item_id = i.id OR x.to_item_id = i.id),
  'editions', (SELECT coalesce(jsonb_agg(to_jsonb(x)), '[]') FROM editions x WHERE x.work_id = i.id),
  'affiliate_links', (SELECT coalesce(jsonb_agg(to_jsonb(x)), '[]') FROM affiliate_links x WHERE x.item_id = i.id),
  -- user-owned rows are counted, never copied: the audit is append-only and
  -- must not outlive account erasure (LGPD, SDD 03 §5)
  'ratings_count', (SELECT count(*) FROM ratings x WHERE x.item_id = i.id),
  'item_status_count', (SELECT count(*) FROM item_status x WHERE x.item_id = i.id),
  'time_to_beat_count', (SELECT count(*) FROM time_to_beat x WHERE x.item_id = i.id)
) FROM items i WHERE i.id = $1`

// moveByKey moves rows of table from the source item to the target, skipping
// keys the target already has (target wins, SDD 01 §7.5). Table, column and
// key names are constants, never user input.
func moveByKey(ctx context.Context, ex database.Executor, from, into, table, itemCol string, keys ...string) error {
	var match strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&match, " AND x.%s = t.%s", k, k)
	}

	_, err := ex.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %[1]s t SET %[2]s = $2 WHERE t.%[2]s = $1
AND NOT EXISTS (SELECT 1 FROM %[1]s x WHERE x.%[2]s = $2%[3]s)`, table, itemCol, match.String()), from, into)

	return err
}

// Merge folds item from into item into (same type) in one transaction: it
// snapshots both into catalog_merge_audit, moves translations, alt titles,
// entities, links, editions, affiliate links and user data (target wins on
// conflicts, except ratings, statuses and time-to-beat where the newer row
// wins), unions external_ids, leaves `merge` redirects for the id and slug,
// deletes the source and dispatches `catalog.merged`. It is not reversible
// (open issue 04-9); the snapshot allows a manual restore. actorID may be
// zero for console runs.
func Merge(ctx context.Context, db *database.DB, q *queue.Queue, from, into string, actorID int64) error {
	if from == into {
		return ErrMergeSelf
	}

	return db.Tx(ctx, func(tx *database.Tx) error {
		return merge(ctx, tx, q, from, into, actorID)
	})
}

type mergeRow struct {
	typ, slug string
	primary   sql.NullString
}

func lockMergeRows(ctx context.Context, tx *database.Tx, from, into string) (src, dst mergeRow, err error) {
	rows, err := tx.QueryContext(ctx, `SELECT id::text, type, slug, primary_edition_id::text FROM items
WHERE id IN ($1, $2) ORDER BY id FOR UPDATE`, from, into)
	if err != nil {
		return src, dst, err
	}

	defer func() { _ = rows.Close() }()

	found := map[string]mergeRow{}

	for rows.Next() {
		var id string

		var r mergeRow
		if err := rows.Scan(&id, &r.typ, &r.slug, &r.primary); err != nil {
			return src, dst, err
		}

		found[id] = r
	}

	if err := rows.Err(); err != nil {
		return src, dst, err
	}

	src, okSrc := found[from]
	dst, okDst := found[into]

	if !okDst {
		return src, dst, ErrMergeNotFound
	}

	if !okSrc {
		if _, merged, err := ResolveID(ctx, tx, ScopeItem, dst.typ, from); err != nil {
			return src, dst, err
		} else if merged {
			return src, dst, ErrMergeRedirected
		}

		return src, dst, ErrMergeNotFound
	}

	if src.typ != dst.typ {
		return src, dst, ErrMergeTypeMismatch
	}

	return src, dst, nil
}

func merge(ctx context.Context, tx *database.Tx, q *queue.Queue, from, into string, actorID int64) error {
	src, dst, err := lockMergeRows(ctx, tx, from, into)
	if err != nil {
		return err
	}

	var snapFrom, snapInto string
	if err := tx.QueryRowContext(ctx, snapshotSQL, from).Scan(&snapFrom); err != nil {
		return fmt.Errorf("snapshot source: %w", err)
	}

	if err := tx.QueryRowContext(ctx, snapshotSQL, into).Scan(&snapInto); err != nil {
		return fmt.Errorf("snapshot target: %w", err)
	}

	var actor sql.NullInt64
	if actorID != 0 {
		actor = sql.NullInt64{Int64: actorID, Valid: true}
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_merge_audit
(actor_id, item_type, source_item_id, target_item_id, snapshot)
VALUES ($1, $2, $3, $4, jsonb_build_object('source', $5::jsonb, 'target', $6::jsonb))`,
		actor, src.typ, from, into, snapFrom, snapInto); err != nil {
		return fmt.Errorf("write audit: %w", err)
	}

	if err := moveRelations(ctx, tx, from, into); err != nil {
		return err
	}

	if err := moveUserData(ctx, tx, from, into); err != nil {
		return err
	}

	// Primary edition: keep the target's, else adopt the source's (already moved).
	if _, err := tx.ExecContext(ctx, `UPDATE items SET
external_ids = (SELECT external_ids FROM items WHERE id = $3) || external_ids,
primary_edition_id = coalesce(primary_edition_id, $4::uuid),
updated_at = now() WHERE id = $1 AND type = $2`,
		into, dst.typ, from, src.primary); err != nil {
		return fmt.Errorf("update target: %w", err)
	}

	if err := RecordMerge(ctx, tx, ScopeItem, src.typ, Target{ID: from, Slug: src.slug}, into); err != nil {
		return fmt.Errorf("record redirects: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM items WHERE id = $1`, from); err != nil {
		return fmt.Errorf("delete source: %w", err)
	}

	return q.DispatchTx(ctx, tx, Merged{Type: src.typ, FromID: from, IntoID: into})
}

func moveRelations(ctx context.Context, ex database.Executor, from, into string) error {
	// Links between the two items would become self links.
	if _, err := ex.ExecContext(ctx, `DELETE FROM item_links
WHERE (from_item_id = $1 AND to_item_id = $2) OR (from_item_id = $2 AND to_item_id = $1)`, from, into); err != nil {
		return fmt.Errorf("drop links between items: %w", err)
	}

	moves := []struct {
		table, col string
		keys       []string
	}{
		{"item_translations", "item_id", []string{"locale"}},
		{"item_alt_titles", "item_id", []string{"title"}},
		{"item_entities", "item_id", []string{"role", "entity_id"}},
		{"affiliate_links", "item_id", []string{"store"}},
		{"item_links", "from_item_id", []string{"relation", "to_item_id"}},
		{"item_links", "to_item_id", []string{"relation", "from_item_id"}},
	}
	for _, m := range moves {
		if err := moveByKey(ctx, ex, from, into, m.table, m.col, m.keys...); err != nil {
			return fmt.Errorf("move %s: %w", m.table, err)
		}
	}

	// Editions have no conflicting key (isbn13 is globally unique).
	if _, err := ex.ExecContext(ctx, `UPDATE editions SET work_id = $2, updated_at = now() WHERE work_id = $1`,
		from, into); err != nil {
		return fmt.Errorf("move editions: %w", err)
	}

	return nil
}

// moveUserData moves ratings, statuses, time-to-beat reports and rating flags.
// On a (user, item) conflict the newer row wins for the first three; for flags
// the stricter one wins. Scores are recomputed by the `catalog.merged` subscribers.
func moveUserData(ctx context.Context, ex database.Executor, from, into string) error {
	stmts := []string{
		`UPDATE ratings r SET c1 = s.c1, c2 = s.c2, c3 = s.c3, c4 = s.c4, c5 = s.c5, c6 = s.c6,
updated_at = s.updated_at FROM ratings s
WHERE s.item_id = $1 AND r.item_id = $2 AND r.user_id = s.user_id AND s.updated_at > r.updated_at`,
		`UPDATE item_status r SET status = s.status, updated_at = s.updated_at FROM item_status s
WHERE s.item_id = $1 AND r.item_id = $2 AND r.user_id = s.user_id AND s.updated_at > r.updated_at`,
		`UPDATE time_to_beat r SET main_minutes = s.main_minutes, extras_minutes = s.extras_minutes,
completionist_minutes = s.completionist_minutes, updated_at = s.updated_at FROM time_to_beat s
WHERE s.item_id = $1 AND r.item_id = $2 AND r.user_id = s.user_id AND s.updated_at > r.updated_at`,
	}
	for _, q := range stmts {
		if _, err := ex.ExecContext(ctx, q, from, into); err != nil {
			return fmt.Errorf("resolve user data conflicts: %w", err)
		}
	}

	// Anti-abuse state must survive a merge: on a (user, item) conflict the
	// stricter flag (lower shadow weight) wins, limiters keep the later end
	// and open bursts follow the item (SDD 02 §7.5).
	guards := []string{
		`UPDATE rating_flags r SET shadow_weight = s.shadow_weight, reason = s.reason, burst_id = s.burst_id
FROM rating_flags s
WHERE s.item_id = $1 AND r.item_id = $2 AND r.user_id = s.user_id AND s.shadow_weight < r.shadow_weight`,
		`UPDATE items t SET
ratings_limited_until = GREATEST(t.ratings_limited_until, f.ratings_limited_until),
unusual_activity_until = GREATEST(t.unusual_activity_until, f.unusual_activity_until)
FROM items f WHERE f.id = $1 AND t.id = $2`,
		`UPDATE score_bursts SET item_id = $2 WHERE item_id = $1`,
	}
	for _, q := range guards {
		if _, err := ex.ExecContext(ctx, q, from, into); err != nil {
			return fmt.Errorf("carry anti-abuse state: %w", err)
		}
	}

	for _, t := range []string{"ratings", "item_status", "time_to_beat", "rating_flags"} {
		if err := moveByKey(ctx, ex, from, into, t, "item_id", "user_id"); err != nil {
			return fmt.Errorf("move %s: %w", t, err)
		}
	}

	if _, err := ex.ExecContext(
		ctx,
		`UPDATE rating_events SET item_id = $2 WHERE item_id = $1`,
		from,
		into,
	); err != nil {
		return fmt.Errorf("move rating_events: %w", err)
	}

	return nil
}
