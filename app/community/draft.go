package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"
)

// ErrDraftNotFound covers a missing suggestion and one the caller does not own, so the route answers 404 either
// way (SDD 04 §6).
var ErrDraftNotFound = errors.New("community: draft not found")

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Target names what a suggestion edits. Kind is item, edition or entity; ID is empty for a create request.
type Target struct {
	Kind string
	ID   string
}

type storedChange struct {
	Before json.RawMessage `json:"before,omitempty"`
	After  json.RawMessage `json:"after,omitempty"`
	Add    []RelationOp    `json:"add,omitempty"`
	Remove []RelationOp    `json:"remove,omitempty"`
}

// SaveDraft validates in and stores it as a draft owned by authorID: id == "" creates one at version 1, otherwise
// the draft is updated when version still matches. It returns the id and the new version. A draft that is missing
// or owned by someone else is ErrDraftNotFound; one past `draft` or at another version is a *TransitionError.
// Eligibility and the daily quota are enforced at submit, not here (SDD 04 §7.3).
func SaveDraft(
	ctx context.Context,
	db *database.DB,
	authorID int64,
	id string,
	version int,
	target Target,
	in SuggestionInput,
	now time.Time,
) (string, int, error) {
	if err := Validate(in, now); err != nil {
		return "", 0, err
	}

	if err := checkTarget(target, in.Kind); err != nil {
		return "", 0, err
	}

	if in.Kind == "create" {
		if err := checkCreate(target, in); err != nil {
			return "", 0, err
		}
	}

	changes := make(map[string]storedChange, len(in.Changes))
	for field, c := range in.Changes {
		changes[field] = storedChange(c)
	}

	raw, err := json.Marshal(changes)
	if err != nil {
		return "", 0, fmt.Errorf("community: encode changes: %w", err)
	}

	urls := in.SourceURLs
	if urls == nil {
		urls = []string{}
	}

	if id == "" {
		return createDraft(ctx, db, authorID, target, in, raw, urls)
	}

	// A malformed id would be a uuid cast error (500) in the query; it is simply not a draft of this author.
	if !uuidRe.MatchString(id) {
		return "", 0, ErrDraftNotFound
	}

	var next int

	err = db.Tx(ctx, func(tx *database.Tx) error {
		var owner sql.NullInt64
		var status string
		var current int

		err := tx.QueryRowContext(ctx,
			`SELECT author_id, status, version FROM suggestions WHERE id = $1 AND kind = $2 FOR UPDATE`,
			id, in.Kind).Scan(&owner, &status, &current)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && (!owner.Valid || owner.Int64 != authorID)) {
			return ErrDraftNotFound
		}

		if err != nil {
			return fmt.Errorf("community: lock draft: %w", err)
		}

		if status != "draft" {
			return &TransitionError{Status: 409, Code: "invalid-transition", From: status, To: "draft"}
		}

		if current != version {
			return stale(status, "draft", current)
		}

		next = current + 1
		_, err = tx.ExecContext(ctx, `UPDATE suggestions
SET locale = $2, changes = $3::jsonb, note = $4, source_urls = $5, version = $6, updated_at = $7
WHERE id = $1`, id, in.Locale, raw, in.Note, urls, next, now.UTC())
		if err != nil {
			return fmt.Errorf("community: update draft: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", 0, err
	}

	return id, next, nil
}

// DraftTarget returns the target and domain stored on authorID's suggestion id. An update must use these, never
// the request's: the target is fixed at create. A missing, malformed or foreign id is ErrDraftNotFound.
func DraftTarget(ctx context.Context, db *database.DB, authorID int64, id string) (Target, string, error) {
	if !uuidRe.MatchString(id) {
		return Target{}, "", ErrDraftNotFound
	}

	var t Target
	var targetID sql.NullString
	var domain string

	err := db.QueryRowContext(ctx,
		`SELECT target_type::text, target_id::text, target_domain FROM suggestions WHERE id = $1 AND author_id = $2`,
		id, authorID).Scan(&t.Kind, &targetID, &domain)
	if errors.Is(err, sql.ErrNoRows) {
		return Target{}, "", ErrDraftNotFound
	}

	if err != nil {
		return Target{}, "", fmt.Errorf("community: draft target: %w", err)
	}

	t.ID = targetID.String

	return t, domain, nil
}

func checkTarget(t Target, kind string) error {
	switch t.Kind {
	case "item", "edition", "entity":
	default:
		return invalid("target_type", "invalid")
	}

	if kind == "create" {
		if t.ID != "" {
			return invalid("target_id", "unexpected")
		}

		return nil
	}

	if !uuidRe.MatchString(t.ID) {
		return invalid("target_id", "invalid")
	}

	return nil
}

// createOnly are the fields that shape a new entity or edition at insert; they are never edited afterwards.
var createOnly = map[string][]string{"entity": {"entity_kind"}, "edition": {"work_id", "language"}}

// checkCreate refuses a create whose fields do not fit its target kind, so the apply never meets a field the
// catalogue cannot store (SDD 04 section 7.6): an entity needs its kind, an edition its book and language and one
// external id (the editions CHECK).
func checkCreate(t Target, in SuggestionInput) error {
	if t.Kind == "edition" && in.TargetType != "book" {
		return invalid("target_type", "editions belong to books")
	}

	allowed := append(EditableFields(t.Kind, in.TargetType), createOnly[t.Kind]...)
	for field := range in.Changes {
		if !slices.Contains(allowed, field) {
			return invalid(field, "unknown")
		}
	}

	for _, field := range createOnly[t.Kind] {
		if _, ok := in.Changes[field]; !ok {
			return invalid(field, "required")
		}
	}

	if t.Kind == "edition" && !slices.ContainsFunc(EditableFields("edition", ""), func(f string) bool {
		_, ok := in.Changes[f]

		return ok && strings.HasPrefix(f, "external_ids.")
	}) {
		return invalid("external_ids", "required")
	}

	return nil
}

func createDraft(
	ctx context.Context,
	db *database.DB,
	authorID int64,
	t Target,
	in SuggestionInput,
	raw []byte,
	urls []string,
) (string, int, error) {
	id := uuid.NewV7().String()

	var targetID any
	if t.ID != "" {
		targetID = t.ID
	}

	err := db.Tx(ctx, func(tx *database.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, note, source_urls)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10)`,
			id, authorID, in.Kind, t.Kind, targetID, in.TargetType, in.Locale, raw, in.Note, urls)
		if err != nil {
			return fmt.Errorf("community: insert draft: %w", err)
		}

		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO suggestion_events (suggestion_id, from_status, to_status, actor_id, actor_role)
VALUES ($1, NULL, 'draft', $2, 'author')`,
			id,
			authorID,
		)
		if err != nil {
			return fmt.Errorf("community: insert draft event: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", 0, err
	}

	return id, 1, nil
}
