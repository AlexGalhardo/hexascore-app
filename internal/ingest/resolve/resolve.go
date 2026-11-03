// Package resolve is the entity-resolution step of the normaliser (RFC-001 section 5 C): before a new item is created
// it checks that none of its external ids belongs to another item and that no other item has almost the same title and
// year. The worker never merges items; it holds the incoming one for an admin (item_conflicts) or drops the contested id.
package resolve

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// UniqueKeys are the external ids that identify one item. They are interpolated into SQL as literals so the unique
// expression indexes serve the lookups; the list is a constant, never input.
var UniqueKeys = []string{"epic", "gog", "igdb", "imdb", "openlibrary", "steam", "tmdb"}

// The near-duplicate rule: the same normalised English title within a year, trigram similarity at least this.
const (
	MinSimilarity = 0.92
	MaxYearGap    = 1
)

// Conflict kinds and states of item_conflicts.
const (
	KindExternalID = "external_id"
	KindTitleYear  = "title_year"

	StateOpen      = "open"
	StateAllowed   = "allowed"
	StateDismissed = "dismissed"
)

// Input is the incoming item. Existing is true when its Wikidata QID already has an item.
type Input struct {
	QID, Type, Title string
	Year             int
	External         map[string]string
	Existing         bool
}

// Conflict is one finding.
type Conflict struct {
	Kind, Key, Value, Other, State string
	Similarity                     float64
}

// Decision tells the normaliser what to do. Drop lists the external ids to take out of the incoming item. Hold means
// the item must not be created now.
type Decision struct {
	Hold      bool
	Drop      []string
	Conflicts []Conflict
}

// Evaluate checks the incoming item. Every new finding is recorded as an open conflict, once: a row that exists keeps
// its state. An existing item is never held, but an id another item owns is dropped from its update. A new item is
// held while a conflict is open or dismissed; an allowed external-id conflict lets it in without the contested id.
func Evaluate(ctx context.Context, ex database.Executor, in Input) (Decision, error) {
	var d Decision

	for _, key := range UniqueKeys {
		val := in.External[key]
		if val == "" {
			continue
		}

		other, ok, err := owner(ctx, ex, key, val, in)
		if err != nil {
			return d, err
		}

		if !ok {
			continue
		}

		c := Conflict{Kind: KindExternalID, Key: key, Value: val, Other: other}

		state, err := settle(ctx, ex, in, c)
		if err != nil {
			return d, err
		}

		c.State = state
		d.Conflicts = append(d.Conflicts, c)
		d.Drop = append(d.Drop, key)

		if !in.Existing && state != StateAllowed {
			d.Hold = true
		}
	}

	if d.Hold || in.Existing || in.Year == 0 || strings.TrimSpace(in.Title) == "" {
		return d, nil
	}

	other, sim, ok, err := nearDuplicate(ctx, ex, in)
	if err != nil || !ok {
		return d, err
	}

	c := Conflict{Kind: KindTitleYear, Value: in.Title, Other: other, Similarity: sim}

	state, err := settle(ctx, ex, in, c)
	if err != nil {
		return d, err
	}

	c.State = state
	d.Conflicts = append(d.Conflicts, c)
	d.Hold = state != StateAllowed

	return d, nil
}

// owner finds the other public item that holds the external id; a hidden or taken-down item does not own it. The TMDB
// id is scoped to the media type.
func owner(ctx context.Context, ex database.Executor, key, val string, in Input) (string, bool, error) {
	q := fmt.Sprintf(
		`SELECT id::text FROM items WHERE (external_ids->>'%s') = $1 AND wikidata_qid <> $2 AND visibility = 'public'`,
		key,
	)
	args := []any{val, in.QID}

	if key == "tmdb" {
		q += ` AND type = $3`
		args = append(args, in.Type)
	}

	var id string

	err := ex.QueryRowContext(ctx, q+` ORDER BY created_at, id LIMIT 1`, args...).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("resolve: owner of %s %s: %w", key, val, err)
	}

	return id, true, nil
}

// nearDuplicate finds the closest other item of the type with a similar English title and a release year within
// MaxYearGap (the query lives in internal/search, which owns the title matching SQL).
func nearDuplicate(ctx context.Context, ex database.Executor, in Input) (string, float64, bool, error) {
	return search.NearTitle(ctx, ex, in.Type, in.Title, in.QID, in.Year, MaxYearGap, MinSimilarity)
}

// settle returns the state of the conflict, recording it as open when it is new.
func settle(ctx context.Context, ex database.Executor, in Input, c Conflict) (string, error) {
	var state string

	err := ex.QueryRowContext(ctx, `INSERT INTO item_conflicts (wikidata_qid, item_type, kind, key, value, other_item_id, similarity)
VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7::float8, 0))
ON CONFLICT (wikidata_qid, kind, key, other_item_id) DO UPDATE SET value = item_conflicts.value
RETURNING status`, in.QID, in.Type, c.Kind, c.Key, c.Value, c.Other, c.Similarity).
		Scan(&state)
	if err != nil {
		return "", fmt.Errorf("resolve: record conflict of %s: %w", in.QID, err)
	}

	return state, nil
}
