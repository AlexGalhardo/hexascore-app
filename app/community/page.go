package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/flags"
)

// SubmissionsFlag is the submission kill switch of SDD 04 §6, default off. SDD 04 names it community.submissions;
// vukem flag names allow only [a-z0-9_], so the code name is community_submissions (env FLAG_COMMUNITY_SUBMISSIONS).
var SubmissionsFlag = flags.Bool("community_submissions", false).
	Describe("Accept community edit suggestions (SDD 04 §6); keep off until moderation is staffed")

// ErrTargetNotFound is a suggest target that does not exist or is not public.
var ErrTargetNotFound = errors.New("community: target not found")

// Snapshot is the record a suggestion edits. Domain is the item type; for an entity it is the caller's choice.
// Values holds the current value of each editable scalar field, as the validator expects it in `before`.
type Snapshot struct {
	Kind   string                     `json:"kind"`
	ID     string                     `json:"id"`
	Domain string                     `json:"domain"`
	Values map[string]json.RawMessage `json:"values"`
}

// Lock is an active field lock on the target; Field "*" locks the whole record.
type Lock struct {
	Field     string     `json:"field"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// Quota is the author's daily submit allowance (SDD 04 §7.3).
type Quota struct {
	Limit   int       `json:"limit"`
	Used    int       `json:"used"`
	Left    int       `json:"left"`
	ResetAt time.Time `json:"resetAt"`
}

// DraftView is the author's saved draft for a target, to restore into the form.
type DraftView struct {
	ID         string          `json:"id"`
	Version    int             `json:"version"`
	Locale     string          `json:"locale"`
	Note       string          `json:"note"`
	SourceURLs []string        `json:"sourceUrls"`
	Changes    json.RawMessage `json:"changes"`
}

// EditPage is everything Community/SuggestEdit needs.
type EditPage struct {
	Target Snapshot   `json:"target"`
	Fields []string   `json:"fields"`
	Locks  []Lock     `json:"locks"`
	Quota  Quota      `json:"quota"`
	Draft  *DraftView `json:"draft"`
}

// EditableFields lists the fields a suggestion may change on a record (SDD 04 §7.2).
func EditableFields(kind, domain string) []string {
	switch kind {
	case "edition":
		return []string{"title", "cover", "external_ids.isbn13", "external_ids.wikidata", "external_ids.openlibrary"}
	case "entity":
		return []string{"title", "cover", "external_ids.wikidata"}
	}

	fields := []string{
		"title", "release_date", "cover",
		"translations.en.title", "translations.pt-BR.title", "translations.es.title",
		"synopsis.en", "synopsis.pt-BR", "synopsis.es",
		"genres", "creators", "organizations", "external_ids.wikidata",
	}
	switch domain {
	case "movie", "series":
		fields = append(fields, "external_ids.tmdb")
	case "game":
		fields = append(fields, "external_ids.igdb", "franchise", "platforms")
	case "book":
		fields = append(fields, "external_ids.openlibrary")
	}

	return fields
}

// LoadEditPage reads the target, its locks, the author's quota and latest draft for it. A missing or non-public
// target is ErrTargetNotFound. domainHint is used only for entities, which have no domain of their own.
func LoadEditPage(
	ctx context.Context,
	db *database.DB,
	authorID int64,
	t Target,
	domainHint string,
	now time.Time,
) (EditPage, error) {
	page := EditPage{Locks: []Lock{}}

	err := db.Tx(ctx, func(tx *database.Tx) error {
		domain, err := targetDomain(ctx, tx, t, domainHint)
		if err != nil {
			return err
		}

		page.Target = Snapshot{Kind: t.Kind, ID: t.ID, Domain: domain, Values: map[string]json.RawMessage{}}
		page.Fields = EditableFields(t.Kind, domain)

		for _, f := range page.Fields {
			if relationField(f) {
				continue
			}
			v, err := scalarCurrent(ctx, tx, t.Kind, t.ID, f)
			if err != nil {
				return err
			}
			page.Target.Values[f] = v
		}

		if page.Locks, err = activeLocks(ctx, tx, t, now); err != nil {
			return err
		}

		used, err := submittedToday(ctx, tx, authorID, now)
		if err != nil {
			return err
		}
		page.Quota = Quota{
			Limit: dailyLimit, Used: used, Left: max(dailyLimit-used, 0), ResetAt: nextUTCMidnight(now),
		}

		page.Draft, err = latestDraft(ctx, tx, authorID, t)

		return err
	})

	return page, err
}

func relationField(f string) bool {
	switch f {
	case "genres", "creators", "organizations", "franchise", "platforms":
		return true
	}

	return false
}

// TargetDomain resolves the domain (item type) a draft on t belongs to, or ErrTargetNotFound.
func TargetDomain(ctx context.Context, db *database.DB, t Target, domainHint string) (string, error) {
	var domain string
	err := db.Tx(ctx, func(tx *database.Tx) (err error) {
		domain, err = targetDomain(ctx, tx, t, domainHint)

		return err
	})

	return domain, err
}

func targetDomain(ctx context.Context, tx *database.Tx, t Target, hint string) (string, error) {
	if !uuidRe.MatchString(t.ID) {
		return "", ErrTargetNotFound
	}

	var q string
	switch t.Kind {
	case "item":
		q = `SELECT type FROM items WHERE id = $1 AND visibility = 'public'`
	case "edition":
		q = `SELECT i.type FROM editions e JOIN items i ON i.id = e.work_id WHERE e.id = $1 AND i.visibility = 'public'`
	case "entity":
		q = `SELECT $2::text FROM entities WHERE id = $1`
	default:
		return "", ErrTargetNotFound
	}

	args := []any{t.ID}
	if t.Kind == "entity" {
		args = append(args, hint)
	}

	var domain string
	err := tx.QueryRowContext(ctx, q, args...).Scan(&domain)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTargetNotFound
	}
	if err != nil {
		return "", fmt.Errorf("community: read target: %w", err)
	}

	return domain, nil
}

func activeLocks(ctx context.Context, tx *database.Tx, t Target, now time.Time) ([]Lock, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT field, reason, expires_at FROM catalog_locks
WHERE target_type = $1 AND target_id = $2 AND (expires_at IS NULL OR expires_at > $3)
ORDER BY field`, t.Kind, t.ID, now.UTC())
	if err != nil {
		return nil, fmt.Errorf("community: read locks: %w", err)
	}
	defer rows.Close()

	locks := []Lock{}
	for rows.Next() {
		var l Lock
		var exp sql.NullTime
		if err := rows.Scan(&l.Field, &l.Reason, &exp); err != nil {
			return nil, fmt.Errorf("community: read locks: %w", err)
		}
		if exp.Valid {
			l.ExpiresAt = &exp.Time
		}
		locks = append(locks, l)
	}

	return locks, rows.Err()
}

func latestDraft(ctx context.Context, tx *database.Tx, authorID int64, t Target) (*DraftView, error) {
	return latestDraftOf(ctx, tx, authorID, "edit", t.Kind, t.ID, "")
}

// latestDraftOf finds the newest draft of kind: an edit by target, a create (no target) or merge by domain.
func latestDraftOf(
	ctx context.Context,
	tx *database.Tx,
	authorID int64,
	kind, targetType, targetID, domain string,
) (*DraftView, error) {
	var d DraftView
	var urls []string
	err := tx.QueryRowContext(ctx, `
SELECT id, version, locale, note, source_urls, changes FROM suggestions
WHERE author_id = $1 AND status = 'draft' AND kind = $2 AND target_type = $3
AND ($4 = '' OR target_id::text = $4) AND ($5 = '' OR target_domain = $5)
ORDER BY updated_at DESC LIMIT 1`, authorID, kind, targetType, targetID, domain).
		Scan(&d.ID, &d.Version, &d.Locale, &d.Note, &urls, &d.Changes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil //nolint:nilnil // no draft is a normal answer
	}
	if err != nil {
		return nil, fmt.Errorf("community: read draft: %w", err)
	}
	d.SourceURLs = urls
	if d.SourceURLs == nil {
		d.SourceURLs = []string{}
	}

	return &d, nil
}

// CreatePage is everything Community/SuggestCreate needs: a new-record (create) or duplicate-report (merge) form.
type CreatePage struct {
	Domain string     `json:"domain"`
	Kind   string     `json:"kind"`
	Fields []string   `json:"fields"`
	Quota  Quota      `json:"quota"`
	Draft  *DraftView `json:"draft"`
}

// LoadCreatePage reads the quota and latest draft of kind ("create" or "merge") for domain. An unknown domain or
// kind is ErrTargetNotFound.
func LoadCreatePage(
	ctx context.Context,
	db *database.DB,
	authorID int64,
	domain, kind string,
	now time.Time,
) (CreatePage, error) {
	if _, err := targetSchema.Parse(domain); err != nil || (kind != "create" && kind != "merge") {
		return CreatePage{}, ErrTargetNotFound
	}

	page := CreatePage{Domain: domain, Kind: kind, Fields: EditableFields("item", domain)}

	err := db.Tx(ctx, func(tx *database.Tx) error {
		used, err := submittedToday(ctx, tx, authorID, now)
		if err != nil {
			return err
		}
		page.Quota = Quota{
			Limit: dailyLimit, Used: used, Left: max(dailyLimit-used, 0), ResetAt: nextUTCMidnight(now),
		}

		page.Draft, err = latestDraftOf(ctx, tx, authorID, kind, "item", "", domain)

		return err
	})

	return page, err
}
