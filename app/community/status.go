package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// StatusView is everything Community/SuggestionStatus shows. Changes is the stored before/after (or add/remove)
// per field; the page renders it as a text diff.
type StatusView struct {
	ID           string            `json:"id"`
	AuthorID     sql.NullInt64     `json:"-"`
	Kind         string            `json:"kind"`
	TargetType   string            `json:"targetType"`
	TargetID     *string           `json:"targetId"`
	Domain       string            `json:"domain"`
	Status       string            `json:"status"`
	Version      int               `json:"version"`
	Locale       string            `json:"locale"`
	Note         string            `json:"note"`
	SourceURLs   []string          `json:"sourceUrls"`
	Changes      json.RawMessage   `json:"changes"`
	Names        map[string]string `json:"names"`
	RejectReason *string           `json:"rejectReason"`
	DecisionNote *string           `json:"decisionNote"`
	SubmittedAt  *time.Time        `json:"submittedAt"`
	CreatedAt    time.Time         `json:"createdAt"`
	Events       []StatusEvent     `json:"events"`
}

// StatusEvent is one timeline entry.
type StatusEvent struct {
	From      *string   `json:"from"`
	To        string    `json:"to"`
	ActorRole string    `json:"actorRole"`
	Reason    *string   `json:"reason"`
	At        time.Time `json:"at"`
}

// LoadStatus reads one suggestion with its timeline. A malformed or unknown id is ErrDraftNotFound; who may see it
// is the caller's policy decision (it needs AuthorID).
func LoadStatus(ctx context.Context, db *database.DB, id string) (StatusView, error) {
	if !uuidRe.MatchString(id) {
		return StatusView{}, ErrDraftNotFound
	}

	v := StatusView{Events: []StatusEvent{}, Names: map[string]string{}}
	var targetID, reject, decision, note sql.NullString
	var submitted sql.NullTime
	var urls []string

	err := db.QueryRowContext(ctx, `
SELECT id, author_id, kind::text, target_type::text, target_id::text, target_domain, status::text, version, locale,
 note, source_urls, changes, reject_reason::text, decision_note, submitted_at, created_at
FROM suggestions WHERE id = $1`, id).Scan(&v.ID, &v.AuthorID, &v.Kind, &v.TargetType, &targetID, &v.Domain,
		&v.Status, &v.Version, &v.Locale, &note, &urls, &v.Changes, &reject, &decision, &submitted, &v.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return StatusView{}, ErrDraftNotFound
	}
	if err != nil {
		return StatusView{}, fmt.Errorf("community: load status: %w", err)
	}

	v.Note, v.SourceURLs = note.String, urls
	if v.SourceURLs == nil {
		v.SourceURLs = []string{}
	}
	v.TargetID, v.RejectReason, v.DecisionNote = ptr(targetID), ptr(reject), ptr(decision)
	if submitted.Valid {
		v.SubmittedAt = &submitted.Time
	}

	if v.Names, err = changeNames(ctx, db, v.Changes); err != nil {
		return StatusView{}, err
	}

	rows, err := db.QueryContext(ctx, `
SELECT from_status::text, to_status::text, actor_role::text, reason, created_at
FROM suggestion_events WHERE suggestion_id = $1 ORDER BY id`, id)
	if err != nil {
		return StatusView{}, fmt.Errorf("community: load events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var e StatusEvent
		var from, reason sql.NullString
		if err := rows.Scan(&from, &e.To, &e.ActorRole, &reason, &e.At); err != nil {
			return StatusView{}, fmt.Errorf("community: load events: %w", err)
		}
		e.From, e.Reason = ptr(from), ptr(reason)
		v.Events = append(v.Events, e)
	}

	return v, rows.Err()
}

// changeNames maps the entity and item ids a diff mentions (relation ops, merge pair) to display names, so the page
// shows names instead of raw ids. Unknown ids are left out.
func changeNames(ctx context.Context, db *database.DB, raw json.RawMessage) (map[string]string, error) {
	var changes map[string]struct {
		Before json.RawMessage `json:"before"`
		After  json.RawMessage `json:"after"`
		Add    []RelationOp    `json:"add"`
		Remove []RelationOp    `json:"remove"`
	}
	names := map[string]string{}
	if json.Unmarshal(raw, &changes) != nil {
		return names, nil
	}

	seen := map[string]bool{}
	for _, c := range changes {
		cands := []string{}
		for _, op := range append(slices.Clone(c.Add), c.Remove...) {
			cands = append(cands, op.ID)
		}
		for _, r := range []json.RawMessage{c.Before, c.After} {
			var id string
			if json.Unmarshal(r, &id) == nil {
				cands = append(cands, id)
			}
		}
		for _, id := range cands {
			seen[id] = uuidRe.MatchString(id)
		}
	}

	ids := []any{}
	marks := []string{}
	for id, ok := range seen {
		if ok {
			ids = append(ids, id)
			marks = append(marks, fmt.Sprintf("$%d::uuid", len(ids)))
		}
	}
	if len(ids) == 0 {
		return names, nil
	}

	// Compare the uuid column (not id::text) so the primary key index serves the lookup.
	in := strings.Join(marks, ",")
	rows, err := db.QueryContext(ctx, `
SELECT id::text, default_name FROM entities WHERE id IN (`+in+`)
UNION ALL
SELECT id::text, original_title FROM items WHERE original_title IS NOT NULL AND id IN (`+in+`)`, ids...)
	if err != nil {
		return nil, fmt.Errorf("community: load change names: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("community: load change names: %w", err)
		}
		names[id] = name
	}

	return names, rows.Err()
}

func ptr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}

	return &s.String
}

// QuotaFor is the author's daily submit allowance at now.
func QuotaFor(ctx context.Context, db *database.DB, authorID int64, now time.Time) (Quota, error) {
	var used int
	err := db.Tx(ctx, func(tx *database.Tx) (err error) {
		used, err = submittedToday(ctx, tx, authorID, now)

		return err
	})
	if err != nil {
		return Quota{}, err
	}

	return Quota{Limit: dailyLimit, Used: used, Left: max(dailyLimit-used, 0), ResetAt: nextUTCMidnight(now)}, nil
}

// MineItem is one row of Community/MySuggestions.
type MineItem struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	Title     *string   `json:"title"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListMine returns the author's newest suggestions (at most limit), withdrawn and expired included.
func ListMine(ctx context.Context, db *database.DB, authorID int64, limit int) ([]MineItem, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, kind::text, target_domain, status::text, changes #>> '{title,after}', updated_at
FROM suggestions WHERE author_id = $1 ORDER BY updated_at DESC, id DESC LIMIT $2`, authorID, limit)
	if err != nil {
		return nil, fmt.Errorf("community: list suggestions: %w", err)
	}
	defer rows.Close()

	items := []MineItem{}
	for rows.Next() {
		var m MineItem
		var title sql.NullString
		if err := rows.Scan(&m.ID, &m.Kind, &m.Domain, &m.Status, &title, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("community: list suggestions: %w", err)
		}
		m.Title = ptr(title)
		items = append(items, m)
	}

	return items, rows.Err()
}

// Withdraw lets the author pull a draft, submitted or in-review suggestion back (SDD 04 §7.1). Other statuses are a
// 409 invalid-transition; the caller has already checked ownership.
func Withdraw(ctx context.Context, db *database.DB, id string, authorID int64, version int) error {
	return Transition(ctx, db, id, version, "withdrawn",
		Actor{ID: sql.NullInt64{Int64: authorID, Valid: true}, Role: "author"})
}
