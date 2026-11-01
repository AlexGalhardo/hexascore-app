package dq

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// FrontendJob is the job_runs.job of the missing-field detector.
const FrontendJob = "dq.frontend_fields"

const (
	topSampled    = 200
	randomSampled = 300
	// DailySuggestionCap is the most system "missing field" suggestions filed in a UTC day (SDD 06 §7.7).
	DailySuggestionCap = 200
	missingFieldSource = "dq.frontend_fields"
)

// required lists the fields each type's item page needs; synopsis is measured for the English page.
var required = map[string][]string{
	"game":   {"title", "release_date", "cover", "genre", "synopsis", "developer", "platform"},
	"movie":  {"title", "release_date", "cover", "genre", "synopsis", "runtime_minutes", "director"},
	"series": {"title", "release_date", "cover", "genre", "synopsis", "season_count", "creator_or_network"},
	"book":   {"title", "release_date", "cover", "genre", "synopsis", "author", "edition"},
}

// Suggest files one missing-field suggestion and reports whether a new row was created (community.IngestGate).
type Suggest func(ctx context.Context, tx *database.Tx, itemID, domain, field string) (bool, error)

// FrontendFields is the dq.frontend_fields step: it renders a sample of items through the item page read model,
// records per type and field how many lack a required value in field_coverage, and files at most
// DailySuggestionCap deduplicated system suggestions a day for the misses.
type FrontendFields struct {
	Suggest Suggest
	Now     func() time.Time
}

type sampled struct {
	id, slug string
	seasons  sql.NullInt64
}

// Step runs the detector for one run row.
func (f FrontendFields) Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	now := time.Now
	if f.Now != nil {
		now = f.Now
	}

	day := now().UTC().Truncate(24 * time.Hour)

	var filed int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM suggestions
WHERE author_id IS NULL AND meta->>'source' = $1 AND created_at >= $2`, missingFieldSource, day).Scan(&filed); err != nil {
		return run.Result{}, fmt.Errorf("dq: count today's suggestions: %w", err)
	}

	var res run.Result

	for _, typ := range []string{"game", "movie", "series", "book"} {
		items, err := sample(ctx, tx, typ, r.ID)
		if err != nil {
			return res, err
		}

		missing := map[string]int{}

		for _, it := range items {
			miss, err := missingFields(ctx, tx, typ, it)
			if err != nil {
				return res, err
			}

			for _, field := range miss {
				missing[field]++

				if filed >= DailySuggestionCap {
					continue
				}

				created, err := f.Suggest(ctx, tx, it.id, typ, suggestionField(field))
				if err != nil {
					return res, err
				}

				if created {
					filed++
					res.Changed++
				}
			}
		}

		res.Seen += len(items)

		for _, field := range required[typ] {
			locale := ""
			if field == "synopsis" {
				locale = "en"
			}

			if _, err := tx.ExecContext(ctx, `INSERT INTO field_coverage (run_id, type, field, locale, sampled, missing)
VALUES ($1, $2, $3, $4, $5, $6)`, r.ID, typ, field, locale, len(items), missing[field]); err != nil {
				return res, fmt.Errorf("dq: write field coverage: %w", err)
			}
		}
	}

	res.Done = true

	return res, nil
}

func suggestionField(field string) string {
	if field == "synopsis" {
		return "synopsis.en"
	}

	return field
}

// sample picks the top items by popularity plus a random share of the rest; the order of the random share is
// fixed by the run id so a rerun of the same run sees the same items.
func sample(ctx context.Context, tx *database.Tx, typ string, runID int64) ([]sampled, error) {
	rows, err := tx.QueryContext(ctx, `
WITH top AS (
  SELECT id, slug, season_count FROM items WHERE type = $1 AND visibility = 'public'
  ORDER BY popularity_prior DESC, id LIMIT $2
), rest AS (
  SELECT id, slug, season_count FROM items WHERE type = $1 AND visibility = 'public'
    AND id NOT IN (SELECT id FROM top)
  ORDER BY md5(id::text || $4::text) LIMIT $3
)
SELECT id::text, slug, season_count FROM top UNION ALL SELECT id::text, slug, season_count FROM rest`,
		typ, topSampled, randomSampled, strconv.FormatInt(runID, 10))
	if err != nil {
		return nil, fmt.Errorf("dq: sample %s: %w", typ, err)
	}
	defer func() { _ = rows.Close() }()

	var out []sampled

	for rows.Next() {
		var s sampled
		if err := rows.Scan(&s.id, &s.slug, &s.seasons); err != nil {
			return nil, fmt.Errorf("dq: sample %s: %w", typ, err)
		}

		out = append(out, s)
	}

	return out, rows.Err()
}

func missingFields(ctx context.Context, tx *database.Tx, typ string, it sampled) ([]string, error) {
	page, err := catalog.LoadItemPage(ctx, tx, typ, it.slug, "en")
	if errors.Is(err, catalog.ErrItemNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("dq: load %s: %w", it.slug, err)
	}

	has := map[string]bool{
		"title":              page.Item.Title != "" && page.Item.Title != it.slug,
		"release_date":       page.Item.ReleaseDate != "",
		"cover":              page.Cover != nil,
		"genre":              len(page.Entities["genre"]) > 0,
		"synopsis":           page.Item.Synopsis != nil,
		"runtime_minutes":    page.Item.RuntimeMinutes != nil,
		"season_count":       it.seasons.Valid && it.seasons.Int64 > 0,
		"developer":          len(page.Entities["developer"]) > 0,
		"platform":           len(page.Entities["platform"]) > 0,
		"director":           len(page.Entities["director"]) > 0,
		"creator_or_network": len(page.Entities["creator"])+len(page.Entities["network"]) > 0,
		"author":             len(page.Entities["author"]) > 0,
		"edition":            page.EditionCount > 0,
	}

	var out []string

	for _, field := range required[typ] {
		if !has[field] {
			out = append(out, field)
		}
	}

	return out, nil
}
