package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

// EditionPageSize is how many editions the item page lists inline.
const EditionPageSize = 10

// ttbMinReports is the legacy floor for showing a time-to-beat median (SDD 02 §7.7).
const ttbMinReports = 3

// ErrItemNotFound means no public item matches the type and slug.
var ErrItemNotFound = errors.New("catalog: item not found")

// RatingSummary is the public aggregate; score and categories are 0-10 stars
// and null below the read gate (SDD 02 §6), with the site-only burst fields.
type RatingSummary struct {
	Score               *float64           `json:"score"`
	ScoreByCategory     map[string]float64 `json:"scoreByCategory"`
	RatingCount         int                `json:"ratingCount"`
	Confidence          *string            `json:"confidence"`
	FormulaVersion      *int               `json:"formulaVersion"`
	UnusualActivity     bool               `json:"unusualActivity"`
	RatingsLimitedUntil *time.Time         `json:"ratingsLimitedUntil"`
}

// EntityLink is an entity related to an item, grouped by role on the page.
type EntityLink struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	Slug      string  `json:"slug"`
	Name      string  `json:"name"`
	Character *string `json:"character,omitempty"`
}

// Edition is one book edition.
type Edition struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Subtitle        *string `json:"subtitle"`
	Language        string  `json:"language"`
	Publisher       *string `json:"publisher"`
	PublicationYear *int    `json:"publicationYear"`
	PageCount       *int    `json:"pageCount"`
	Format          *string `json:"format"`
	ISBN13          *string `json:"isbn13"`
}

// TimeToBeat holds the community medians in hours; a style is null below
// three reports (legacy parity).
type TimeToBeat struct {
	Main          *float64 `json:"mainHours"`
	Extras        *float64 `json:"extrasHours"`
	Completionist *float64 `json:"completionistHours"`
	Reports       int      `json:"reports"`
}

// ItemPage is everything Catalog/Show needs that is not user specific.
type ItemPage struct {
	Item          ItemDetail              `json:"item"`
	RatingSummary RatingSummary           `json:"ratingSummary"`
	Entities      map[string][]EntityLink `json:"entities"`
	Editions      []Edition               `json:"editions,omitempty"`
	EditionCount  int                     `json:"editionCount,omitempty"`
	TimeToBeat    *TimeToBeat             `json:"timeToBeat,omitempty"`
}

// LoadItemPage reads the public item of typ and slug in locale, falling back
// to English then to the original title.
func LoadItemPage(ctx context.Context, db Querier, typ, slug, locale string) (ItemPage, error) {
	var (
		page    ItemPage
		src     = ItemSource{Type: typ, Slug: slug, ExternalIDs: map[string]string{}}
		title   sql.NullString
		tsrc    sql.NullString
		syn     sql.NullString
		synSrc  sql.NullString
		synLic  sql.NullString
		synURL  sql.NullString
		release sql.NullTime
		runtime sql.NullInt64
		pages   sql.NullInt64
		limited sql.NullTime
		unusual sql.NullTime
		sum     = &page.RatingSummary
		score   sql.NullFloat64
		cats    [6]sql.NullFloat64
		conf    sql.NullString
		version sql.NullInt64
	)

	err := db.QueryRowContext(ctx, `SELECT i.id::text, i.wikidata_qid, i.release_date, i.release_precision, i.runtime_minutes, i.page_count,
  coalesce(t.title, te.title, i.original_title, i.slug), coalesce(t.title_source, te.title_source, 'hexascore'),
  coalesce(t.synopsis, te.synopsis), coalesce(t.synopsis_source, te.synopsis_source),
  coalesce(t.synopsis_licence, te.synopsis_licence), coalesce(t.synopsis_source_url, te.synopsis_source_url),
  i.rating_count, i.score, i.score_c1, i.score_c2, i.score_c3, i.score_c4, i.score_c5, i.score_c6,
  i.score_confidence, i.score_formula_version, i.ratings_limited_until, i.unusual_activity_until
FROM items i
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $3
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE i.type = $1 AND i.slug = $2 AND i.visibility = 'public'`, typ, slug, locale).
		Scan(
			&src.ID, &src.WikidataQID, &release, &src.ReleasePrecision, &runtime, &pages,
			&title, &tsrc, &syn, &synSrc, &synLic, &synURL,
			&sum.RatingCount, &score, &cats[0], &cats[1], &cats[2], &cats[3], &cats[4], &cats[5],
			&conf, &version, &limited, &unusual)
	if errors.Is(err, sql.ErrNoRows) {
		return page, ErrItemNotFound
	}
	if err != nil {
		return page, fmt.Errorf("load item: %w", err)
	}

	now := time.Now()
	src.Title = Sourced[string]{Value: title.String, Source: tsrc.String}
	if syn.Valid {
		src.Synopsis = &SynopsisIn{
			Text:      syn.String,
			Source:    synSrc.String,
			Licence:   synLic.String,
			SourceURL: synURL.String,
		}
	}
	if release.Valid {
		src.ReleaseDate = &Sourced[time.Time]{Value: release.Time, Source: tsrc.String}
	}
	if runtime.Valid {
		src.RuntimeMinutes = &Sourced[int]{Value: int(runtime.Int64), Source: tsrc.String}
	}
	if pages.Valid {
		src.PageCount = &Sourced[int]{Value: int(pages.Int64), Source: tsrc.String}
	}
	page.Item = Assemble(src, NewPolicy(AudienceSite, nil))

	page.RatingSummary = scoreFields(typ, sum.RatingCount, score, cats, conf, version)
	sum.UnusualActivity = unusual.Valid && unusual.Time.After(now)
	if limited.Valid && limited.Time.After(now) {
		sum.RatingsLimitedUntil = &limited.Time
	}

	if page.Entities, err = loadEntities(ctx, db, src.ID, locale); err != nil {
		return page, err
	}

	switch typ {
	case "book":
		if page.Editions, page.EditionCount, err = loadEditions(ctx, db, src.ID); err != nil {
			return page, err
		}
	case "game":
		if page.TimeToBeat, err = loadTimeToBeat(ctx, db, src.ID); err != nil {
			return page, err
		}
	}

	return page, nil
}

// scoreFields builds the SDD 02 §6 score fields: 0-10 stars with one decimal,
// null below the 5-rating read gate.
func scoreFields(
	typ string, count int, score sql.NullFloat64, cats [6]sql.NullFloat64, conf sql.NullString, version sql.NullInt64,
) RatingSummary {
	stars := func(v float64) float64 { return math.Round(v) / 10 }
	sum := RatingSummary{RatingCount: count, ScoreByCategory: map[string]float64{}}
	if count < 5 || !score.Valid {
		return sum
	}
	s := stars(score.Float64)
	sum.Score = &s
	for i, k := range CategoryKeys[typ] {
		if cats[i].Valid {
			sum.ScoreByCategory[k] = stars(cats[i].Float64)
		}
	}
	if conf.Valid {
		sum.Confidence = &conf.String
	}
	if version.Valid {
		v := int(version.Int64)
		sum.FormulaVersion = &v
	}
	return sum
}

func loadEntities(ctx context.Context, db Querier, itemID, locale string) (map[string][]EntityLink, error) {
	rows, err := db.QueryContext(ctx, `SELECT ie.role, e.id::text, e.kind, e.slug,
  coalesce(t.name, te.name, e.default_name), ie.character_name
FROM item_entities ie
JOIN entities e ON e.id = ie.entity_id
LEFT JOIN entity_translations t ON t.entity_id = e.id AND t.locale = $2
LEFT JOIN entity_translations te ON te.entity_id = e.id AND te.locale = 'en'
WHERE ie.item_id = $1
ORDER BY ie.role, ie.position, e.slug`, itemID, locale)
	if err != nil {
		return nil, fmt.Errorf("load entities: %w", err)
	}
	defer rows.Close()

	out := map[string][]EntityLink{}
	for rows.Next() {
		var (
			role string
			e    EntityLink
			char sql.NullString
		)
		if err := rows.Scan(&role, &e.ID, &e.Kind, &e.Slug, &e.Name, &char); err != nil {
			return nil, fmt.Errorf("scan entity: %w", err)
		}
		if char.Valid {
			e.Character = &char.String
		}
		out[role] = append(out[role], e)
	}

	return out, rows.Err()
}

const editionSelect = `SELECT e.id::text, e.title, e.subtitle, e.language, p.default_name, e.publication_year, e.page_count,
  e.format, e.isbn13
FROM editions e LEFT JOIN entities p ON p.id = e.publisher_id`

func scanEdition(row interface{ Scan(...any) error }) (Edition, error) {
	var (
		e                      Edition
		sub, pub, format, isbn sql.NullString
		year, pages            sql.NullInt64
	)
	if err := row.Scan(&e.ID, &e.Title, &sub, &e.Language, &pub, &year, &pages, &format, &isbn); err != nil {
		return e, err
	}
	if sub.Valid {
		e.Subtitle = &sub.String
	}
	if pub.Valid {
		e.Publisher = &pub.String
	}
	if format.Valid {
		e.Format = &format.String
	}
	if isbn.Valid {
		e.ISBN13 = &isbn.String
	}
	if year.Valid {
		y := int(year.Int64)
		e.PublicationYear = &y
	}
	if pages.Valid {
		p := int(pages.Int64)
		e.PageCount = &p
	}

	return e, nil
}

func loadEditions(ctx context.Context, db Querier, workID string) ([]Edition, int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM editions WHERE work_id = $1`, workID).
		Scan(&count); err != nil {
		return nil, 0, fmt.Errorf("count editions: %w", err)
	}

	rows, err := db.QueryContext(ctx, editionSelect+` WHERE e.work_id = $1
ORDER BY e.is_brazilian DESC, e.publication_year DESC NULLS LAST, e.id LIMIT `+strconv.Itoa(EditionPageSize), workID)
	if err != nil {
		return nil, 0, fmt.Errorf("load editions: %w", err)
	}
	defer rows.Close()

	out := []Edition{}
	for rows.Next() {
		e, err := scanEdition(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan edition: %w", err)
		}
		out = append(out, e)
	}

	return out, count, rows.Err()
}

// LoadEdition reads one edition of the book work; ErrItemNotFound when the
// edition does not belong to it.
func LoadEdition(ctx context.Context, db Querier, workID, editionID string) (Edition, error) {
	e, err := scanEdition(
		db.QueryRowContext(ctx, editionSelect+` WHERE e.work_id = $1 AND e.id::text = $2`, workID, editionID),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return e, ErrItemNotFound
	}

	return e, err
}

func loadTimeToBeat(ctx context.Context, db Querier, itemID string) (*TimeToBeat, error) {
	var (
		t       TimeToBeat
		counts  [3]int
		medians [3]sql.NullFloat64
	)
	err := db.QueryRowContext(ctx, `SELECT count(*), count(main_minutes), count(extras_minutes), count(completionist_minutes),
  percentile_cont(0.5) WITHIN GROUP (ORDER BY main_minutes),
  percentile_cont(0.5) WITHIN GROUP (ORDER BY extras_minutes),
  percentile_cont(0.5) WITHIN GROUP (ORDER BY completionist_minutes)
FROM time_to_beat WHERE item_id = $1`, itemID).
		Scan(
			&t.Reports, &counts[0], &counts[1], &counts[2], &medians[0], &medians[1], &medians[2])
	if err != nil {
		return nil, fmt.Errorf("load time to beat: %w", err)
	}

	hours := func(i int) *float64 {
		if counts[i] < ttbMinReports || !medians[i].Valid {
			return nil
		}
		h := math.Round(medians[i].Float64/60*10) / 10

		return &h
	}
	t.Main, t.Extras, t.Completionist = hours(0), hours(1), hours(2)

	return &t, nil
}

// ListChoice is one of the viewer's lists that accepts the item.
type ListChoice struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	HasItem bool   `json:"hasItem"`
}

// ListsFor returns userID's lists of the item's type, flagging those that already hold it.
func ListsFor(ctx context.Context, db Querier, userID int64, typ, itemID string) ([]ListChoice, error) {
	rows, err := db.QueryContext(ctx, `SELECT l.id::text, l.title,
  EXISTS (SELECT 1 FROM list_items li WHERE li.list_id = l.id AND li.item_id = $3)
FROM lists l WHERE l.user_id = $1 AND l.type = $2 ORDER BY l.updated_at DESC, l.id`, userID, typ, itemID)
	if err != nil {
		return nil, fmt.Errorf("lists for item: %w", err)
	}
	defer rows.Close()

	out := []ListChoice{}
	for rows.Next() {
		var l ListChoice
		if err := rows.Scan(&l.ID, &l.Title, &l.HasItem); err != nil {
			return nil, fmt.Errorf("scan list choice: %w", err)
		}
		out = append(out, l)
	}

	return out, rows.Err()
}
