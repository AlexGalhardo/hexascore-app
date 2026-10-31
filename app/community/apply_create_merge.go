package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// applyCreate inserts the requested item, entity or edition and writes every field through the edit path with
// audit rows source=create and before=null (SDD 04 section 7.6). The new id goes to meta.created_id: target_id
// stays null on a create (table check). The create-only fields (entity kind, parent work, language) shape the
// insert and get no audit row: they are not editable afterwards.
func (a Applier) applyCreate(
	ctx context.Context, tx *database.Tx, s applyRow, id string, fields []string, actor Actor, attempt int,
) error {
	title, _, err := decodeString(s.changes["title"].After)
	if err != nil || strings.TrimSpace(title) == "" {
		return invalid("title", "required")
	}

	if err := catalog.EnsureSource(ctx, tx, catalog.SourceCommunity); err != nil {
		return err
	}

	var newID string

	switch s.targetType {
	case "item":
		newID, err = insertItem(ctx, tx, s, title)
	case "entity":
		newID, err = insertEntity(ctx, tx, s, title)
	case "edition":
		newID, err = insertEdition(ctx, tx, s, title)
	default:
		return &TransitionError{Status: http.StatusConflict, Code: "invalid-transition", From: s.status, To: "applied"}
	}

	if err != nil {
		return err
	}

	s.targetID = newID

	written := slices.DeleteFunc(slices.Clone(fields), func(f string) bool {
		return slices.Contains(createOnly[s.targetType], f)
	})

	auditIDs := make([]int64, 0, len(written))
	for _, field := range written {
		auditID, err := a.applyField(ctx, tx, s, id, field, false, actor, "create")
		if err != nil {
			return err
		}

		auditIDs = append(auditIDs, auditID)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE suggestions SET meta = meta || jsonb_build_object('created_id', $2::text)
WHERE id = $1`, id, newID); err != nil {
		return fmt.Errorf("community: record created id: %w", err)
	}

	if s.targetType == "item" {
		if err := changes.Record(ctx, tx, changes.Change{
			ItemID: newID, ItemType: s.domain, Action: "created", Fields: apiFields(written),
		}); err != nil {
			return err
		}
	}

	return a.finish(ctx, tx, s, id, actor, attempt, CatalogChanged{
		TargetType: s.targetType, TargetID: newID, Fields: written, AuditIDs: auditIDs,
	})
}

func insertItem(ctx context.Context, tx *database.Tx, s applyRow, title string) (string, error) {
	qid, _, err := decodeString(s.changes["external_ids.wikidata"].After)
	if err != nil || qid == "" {
		// items.wikidata_qid is the catalogue's identity (SDD 01), so a create without it cannot be stored.
		return "", invalid("external_ids.wikidata", "required")
	}

	if err := refuseExisting(ctx, tx, `SELECT id::text FROM items WHERE wikidata_qid = $1`, qid); err != nil {
		return "", err
	}

	slug, err := catalog.UniqueSlug(ctx, tx, catalog.ScopeItem, s.domain, title, 0)
	if err != nil {
		return "", invalid("title", "slug")
	}

	newID := uuid.NewV7().String()
	if _, err := tx.ExecContext(ctx, `INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, $2, $3, $4)`,
		newID, s.domain, slug, qid); err != nil {
		return "", duplicateOr(err, "community: insert item")
	}

	return newID, nil
}

func insertEntity(ctx context.Context, tx *database.Tx, s applyRow, title string) (string, error) {
	kind, _, err := decodeString(s.changes["entity_kind"].After)
	if err != nil || kind == "" {
		return "", invalid("entity_kind", "required")
	}

	var qid any

	if q, _, err := decodeString(s.changes["external_ids.wikidata"].After); err == nil && q != "" {
		if err := refuseExisting(ctx, tx, `SELECT id::text FROM entities WHERE wikidata_qid = $1`, q); err != nil {
			return "", err
		}

		qid = q
	}

	slug, err := catalog.UniqueSlug(ctx, tx, catalog.ScopeEntity, kind, title, 0)
	if err != nil {
		return "", invalid("title", "slug")
	}

	newID := uuid.NewV7().String()
	if _, err := tx.ExecContext(ctx, `INSERT INTO entities (id, kind, slug, wikidata_qid, default_name)
VALUES ($1, $2, $3, $4, $5)`, newID, kind, slug, qid, title); err != nil {
		return "", duplicateOr(err, "community: insert entity")
	}

	return newID, nil
}

func insertEdition(ctx context.Context, tx *database.Tx, s applyRow, title string) (string, error) {
	work, _, err := decodeString(s.changes["work_id"].After)
	if err != nil || !uuidRe.MatchString(work) {
		return "", invalid("work_id", "required")
	}

	language, _, err := decodeString(s.changes["language"].After)
	if err != nil || language == "" {
		return "", invalid("language", "required")
	}

	// FOR SHARE keeps the parent work from being merged or deleted under the new edition.
	var found int

	err = tx.QueryRowContext(ctx, `SELECT 1 FROM items WHERE id = $1 AND type = 'book' FOR SHARE`, work).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return "", invalid("work_id", "not found")
	}

	if err != nil {
		return "", fmt.Errorf("community: edition work: %w", err)
	}

	ext := make(map[string]any, len(editionIDs))
	for field, column := range editionIDs {
		ext[column] = nil
		if v, _, err := decodeString(s.changes[field].After); err == nil && v != "" {
			if err := refuseExisting(ctx, tx, `SELECT id::text FROM editions WHERE `+column+` = $1`, v); err != nil {
				return "", err
			}

			ext[column] = v
		}
	}

	newID := uuid.NewV7().String()
	if _, err := tx.ExecContext(ctx, `INSERT INTO editions
(id, work_id, title, language, isbn13, wikidata_qid, openlibrary_id, source)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, newID, work, title, language, ext["isbn13"], ext["wikidata_qid"],
		ext["openlibrary_id"], catalog.SourceCommunity); err != nil {
		return "", duplicateOr(err, "community: insert edition")
	}

	return newID, nil
}

// editionIDs maps the edition external id fields to their unique columns. Values are constants, never input.
var editionIDs = map[string]string{
	"external_ids.isbn13":      "isbn13",
	"external_ids.wikidata":    "wikidata_qid",
	"external_ids.openlibrary": "openlibrary_id",
}

// refuseExisting answers 409 suggestion-duplicate when q (one bound value) finds a record.
func refuseExisting(ctx context.Context, tx *database.Tx, q, value string) error {
	var existing string

	err := tx.QueryRowContext(ctx, q, value).Scan(&existing)
	if err == nil {
		return &Problem{
			Status: http.StatusConflict, Code: "suggestion-duplicate", Type: problemBase + "suggestion-duplicate",
			ExistingID: existing,
		}
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("community: create lookup: %w", err)
	}

	return nil
}

// duplicateOr maps a unique violation (an ingest that won the race on the same identifier) to 409
// suggestion-duplicate; any other error is wrapped with msg.
func duplicateOr(err error, msg string) error {
	if s, ok := errors.AsType[interface {
		error
		SQLState() string
	}](err); ok && s.SQLState() == "23505" {
		return &Problem{
			Status: http.StatusConflict, Code: "suggestion-duplicate", Type: problemBase + "suggestion-duplicate",
		}
	}

	return fmt.Errorf("%s: %w", msg, err)
}

// applyMerge folds the suggestion's target into `*merge*.after` through the SDD 01 merge service, writes one
// `*merge*` audit row with both snapshots, and expires the open suggestions on the deleted item (target_gone). The
// moderator types the deleted item's title: a merge is not revertible.
func (a Applier) applyMerge(
	ctx context.Context, tx *database.Tx, s applyRow, id string, opts ApproveOptions, actor Actor, attempt int,
) error {
	from := s.targetID

	into, _, err := decodeString(s.changes["*merge*"].After)
	if err != nil || into == "" || s.targetType != "item" {
		return invalid("*merge*", "invalid")
	}

	var title string

	err = tx.QueryRowContext(ctx, `SELECT COALESCE(original_title, slug) FROM items WHERE id = $1`, from).Scan(&title)
	if errors.Is(err, sql.ErrNoRows) {
		return &TransitionError{Status: http.StatusConflict, Code: "target-gone", To: "applied"}
	}

	if err != nil {
		return fmt.Errorf("community: merge source title: %w", err)
	}

	if !strings.EqualFold(strings.TrimSpace(opts.ConfirmTitle), strings.TrimSpace(title)) {
		return invalid("confirm_title", "Type the title of the item being merged away.")
	}

	if err := catalog.MergeTx(ctx, tx, a.Queue, from, into, actor.ID.Int64); err != nil {
		if errors.Is(err, catalog.ErrMergeNotFound) || errors.Is(err, catalog.ErrMergeTypeMismatch) ||
			errors.Is(err, catalog.ErrMergeSelf) || errors.Is(err, catalog.ErrMergeRedirected) {
			return invalid("*merge*", err.Error())
		}

		return err
	}

	var snapshot []byte
	if err := tx.QueryRowContext(ctx, `SELECT snapshot FROM catalog_merge_audit WHERE source_item_id = $1
ORDER BY id DESC LIMIT 1`, from).Scan(&snapshot); err != nil {
		return fmt.Errorf("community: merge snapshot: %w", err)
	}

	after, err := json.Marshal(map[string]string{"merged_from": from, "merged_into": into})
	if err != nil {
		return fmt.Errorf("community: merge audit: %w", err)
	}

	var auditID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO catalog_audit
(actor_id, actor_role, source, suggestion_id, target_type, target_id, field, before, after)
VALUES ($1, $2, 'merge', $3, 'item', $4, '*merge*', $5::jsonb, $6::jsonb) RETURNING id`,
		actor.ID, actor.Role, id, into, string(snapshot), string(after)).Scan(&auditID); err != nil {
		return fmt.Errorf("community: merge audit: %w", err)
	}

	if err := a.expireOn(ctx, tx, from, id); err != nil {
		return err
	}

	return a.finish(ctx, tx, s, id, actor, attempt, CatalogChanged{
		TargetType: "item", TargetID: into, Fields: []string{"*merge*"}, AuditIDs: []int64{auditID},
	})
}

// expireOn moves the open suggestions on a deleted target to expired with reason target_gone (SDD 04 section 8),
// one system event each, and queues the author notification. target_gone never touches contributor_stats.
func (a Applier) expireOn(ctx context.Context, tx *database.Tx, targetID, except string) error {
	rows, err := tx.QueryContext(ctx, `WITH o AS (
  SELECT id, status FROM suggestions
  WHERE target_id = $1 AND id <> $2 AND status IN ('submitted','in_review') FOR UPDATE
), u AS (
  UPDATE suggestions s SET status = 'expired', reject_reason = 'target_gone', version = s.version + 1,
    updated_at = now(), assigned_to = NULL, claimed_at = NULL
  FROM o WHERE s.id = o.id RETURNING s.id
)
INSERT INTO suggestion_events (suggestion_id, from_status, to_status, actor_role, reason)
SELECT o.id, o.status, 'expired', 'system', 'target_gone' FROM o JOIN u ON u.id = o.id
RETURNING suggestion_id::text`, targetID, except)
	if err != nil {
		return fmt.Errorf("community: expire suggestions on %s: %w", targetID, err)
	}

	var ids []string

	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			return errors.Join(fmt.Errorf("community: expire suggestions: %w", err), finishRows(rows))
		}

		ids = append(ids, sid)
	}

	if err := finishRows(rows); err != nil {
		return fmt.Errorf("community: expire suggestions: %w", err)
	}

	for _, sid := range ids {
		if err := a.Queue.DispatchTx(ctx, tx, Notify{SuggestionID: sid, State: "expired"}); err != nil {
			return fmt.Errorf("community: dispatch notify: %w", err)
		}
	}

	return nil
}
