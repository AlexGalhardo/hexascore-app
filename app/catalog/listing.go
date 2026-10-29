package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ErrInvalidQuery is returned for an out-of-range or unknown listing
// parameter; handlers answer 400 with problem code invalid-query.
var ErrInvalidQuery = errors.New("invalid-query")

// Listing limits and defaults (SDD 01 FR-01-10).
const (
	DefaultLimit = 20
	MaxLimit     = 100
	DefaultSort  = "-popularity"
)

// CategoryKeys are the six score categories per item type, in column order
// score_c1..score_c6 (legacy lib/categories.ts). Keys double as the public
// `min_<category>` filter names, so renaming one is a breaking change.
var CategoryKeys = map[string][6]string{
	"game":   {"gameplay", "fun", "sound", "technical", "visual", "storytelling"},
	"movie":  {"screenplay", "acting", "direction", "cinematography", "soundtrack", "entertainment"},
	"series": {"screenplay", "acting", "characters", "pacing", "production", "entertainment"},
	"book":   {"plot", "characters", "writing", "pacing", "originality", "impact"},
}

// visibleScore is SDD 02's read gate: scores show from 5 counted ratings.
const visibleScore = "(i.rating_count >= 5 AND i.score IS NOT NULL)"

// sortSpec is one legacy sort field. Each expression embeds a null sentinel
// so nulls sort last in both directions (legacy sortExpression).
type sortSpec struct {
	cast      string
	asc, desc string
}

var sortFields = map[string]sortSpec{
	"score": {
		cast: "numeric",
		asc:  "coalesce(CASE WHEN " + visibleScore + " THEN i.score END, 1000)",
		desc: "coalesce(CASE WHEN " + visibleScore + " THEN i.score END, -1)",
	},
	"release_date": {
		cast: "date",
		asc:  "coalesce(i.release_date, DATE '9999-12-31')",
		desc: "coalesce(i.release_date, DATE '0001-01-01')",
	},
	"rating_count": {cast: "int", asc: "i.rating_count", desc: "i.rating_count"},
	"popularity":   {cast: "real", asc: "i.popularity_prior", desc: "i.popularity_prior"},
}

// ParseSort splits "-field" into field and direction; unknown sorts fail.
func ParseSort(sort string) (field string, desc bool, err error) {
	if sort == "" {
		sort = DefaultSort
	}
	field, desc = strings.CutPrefix(sort, "-")
	if _, ok := sortFields[field]; !ok {
		return "", false, fmt.Errorf("%w: unknown sort %q", ErrInvalidQuery, sort)
	}
	return field, desc, nil
}

// ListParams are the listing inputs; scores are stars (0-10).
type ListParams struct {
	Type            string             // game|movie|series|book ("" only with EntityID)
	EntityID        string             // items related to this entity id (API entity items)
	Role            string             // item_entities role, with EntityID
	RelationSources []string           // non-nil: relation filters match only item_entities rows from these sources (API)
	Genre           string             // genre entity slug
	Platform        string             // platform entity slug
	Year            int                // 0 = any
	MinScore        *float64           // stars
	MaxScore        *float64           // stars
	MinCategory     map[string]float64 // category key -> min stars
	Sort            string             // "[-]popularity|score|release_date|rating_count"
	Limit           int
	Cursor          string
}

// ListedItem is one listing row; summaries join translations elsewhere.
type ListedItem struct{ ID, Type, Slug string }

// ListPage is one page plus signed cursors; Total is set on the first page only.
type ListPage struct {
	Items      []ListedItem
	NextCursor string
	PrevCursor string
	Total      *int
}

// Querier is the subset of *database.DB the listing needs.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func validStars(v float64) bool { return v >= 0 && v <= 10 && v*2 == math.Trunc(v*2) }

type queryBuilder struct {
	where []string
	args  []any
}

// add appends a condition; each "?" becomes the next $n placeholder.
func (b *queryBuilder) add(cond string, args ...any) {
	for _, a := range args {
		b.args = append(b.args, a)
		cond = strings.Replace(cond, "?", fmt.Sprintf("$%d", len(b.args)), 1)
	}
	b.where = append(b.where, cond)
}

func (p ListParams) filters() (*queryBuilder, error) {
	keys, ok := CategoryKeys[p.Type]
	if !ok && (p.Type != "" || p.EntityID == "") {
		return nil, fmt.Errorf("%w: unknown type %q", ErrInvalidQuery, p.Type)
	}
	b := &queryBuilder{}
	if p.Type != "" {
		b.add("i.type = ?", p.Type)
	}
	var srcCond string
	var srcArgs []any
	if p.RelationSources != nil {
		srcCond, srcArgs = " AND ie.source = ANY(?::text[])", []any{"{" + strings.Join(p.RelationSources, ",") + "}"}
	}
	if p.EntityID != "" {
		cond, args := "EXISTS (SELECT 1 FROM item_entities ie WHERE ie.item_id = i.id AND ie.entity_id = ?::uuid"+srcCond,
			append(
				[]any{p.EntityID},
				srcArgs...)
		if p.Role != "" {
			cond, args = cond+" AND ie.role = ?", append(args, p.Role)
		}
		b.add(cond+")", args...)
	}
	b.add("i.visibility = 'public'") // FR-01-14
	entity := func(kind, slug string) {
		b.add(`EXISTS (SELECT 1 FROM item_entities ie JOIN entities e ON e.id = ie.entity_id
WHERE ie.item_id = i.id AND ie.role = ? AND e.kind = ? AND e.slug = ?`+srcCond+")", append([]any{kind, kind, slug}, srcArgs...)...)
	}
	if p.Genre != "" {
		entity("genre", p.Genre)
	}
	if p.Platform != "" {
		entity("platform", p.Platform)
	}
	if p.Year != 0 {
		if p.Year < 1800 || p.Year > 2100 {
			return nil, fmt.Errorf("%w: year out of range", ErrInvalidQuery)
		}
		b.add("i.release_date >= make_date(?, 1, 1) AND i.release_date <= make_date(?, 12, 31)", p.Year, p.Year)
	}
	if p.MinScore != nil {
		if !validStars(*p.MinScore) {
			return nil, fmt.Errorf("%w: min_score must be 0-10 in steps of 0.5", ErrInvalidQuery)
		}
		b.add(visibleScore+" AND i.score >= ?", *p.MinScore*10)
	}
	if p.MaxScore != nil {
		if !validStars(*p.MaxScore) {
			return nil, fmt.Errorf("%w: max_score must be 0-10 in steps of 0.5", ErrInvalidQuery)
		}
		b.add(visibleScore+" AND i.score <= ?", *p.MaxScore*10)
	}
	for key, minStars := range p.MinCategory {
		idx := -1
		for n, k := range keys {
			if k == key {
				idx = n + 1
			}
		}
		if idx < 0 || !validStars(minStars) {
			return nil, fmt.Errorf("%w: min_%s", ErrInvalidQuery, key)
		}
		b.add(visibleScore+fmt.Sprintf(" AND i.score_c%d >= ?", idx), minStars*10)
	}
	return b, nil
}

// List runs a keyset page over (sort expression, id) with the legacy filters.
// Only allowlisted fragments reach SQL; every value is a $n argument.
func List(ctx context.Context, db Querier, codec *CursorCodec, p ListParams) (ListPage, error) {
	if p.Sort == "" {
		p.Sort = DefaultSort
	}
	field, desc, err := ParseSort(p.Sort)
	if err != nil {
		return ListPage{}, err
	}
	if p.Limit == 0 {
		p.Limit = DefaultLimit
	}
	if p.Limit < 1 || p.Limit > MaxLimit {
		return ListPage{}, fmt.Errorf("%w: limit must be 1-%d", ErrInvalidQuery, MaxLimit)
	}
	b, err := p.filters()
	if err != nil {
		return ListPage{}, err
	}

	var total *int
	if p.Cursor == "" {
		var n int
		q := "SELECT count(*) FROM items i WHERE " + strings.Join(b.where, " AND ")
		if err := db.QueryRowContext(ctx, q, b.args...).Scan(&n); err != nil {
			return ListPage{}, fmt.Errorf("count listing: %w", err)
		}
		total = &n
	}

	spec := sortFields[field]
	expr := spec.asc
	if desc {
		expr = spec.desc
	}

	var cur *Cursor
	backwards := false
	if p.Cursor != "" {
		c, err := codec.Decode(p.Cursor, p.Sort)
		if err != nil {
			return ListPage{}, err
		}
		cur, backwards = &c, c.D == DirPrev
	}
	// Walking backwards flips comparison and order; rows are reversed again below.
	op, dir := ">", "ASC"
	if desc != backwards {
		op, dir = "<", "DESC"
	}
	if cur != nil {
		b.add(fmt.Sprintf("(%s, i.id) %s (?::%s, ?::uuid)", expr, op, spec.cast), cur.V, cur.ID)
	}

	q := fmt.Sprintf("SELECT i.id, i.type, i.slug, (%s)::text FROM items i WHERE %s ORDER BY %s %s, i.id %s LIMIT %d",
		expr, strings.Join(b.where, " AND "), expr, dir, dir, p.Limit+1)
	rows, err := db.QueryContext(ctx, q, b.args...)
	if err != nil {
		// The cursor is HMAC-signed, so a value that fails to cast was issued
		// for another column type: report it as invalid, not as a server error.
		if cur != nil {
			return ListPage{}, ErrInvalidCursor
		}
		return ListPage{}, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	type row struct {
		ListedItem
		sortValue string
	}
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.Type, &r.Slug, &r.sortValue); err != nil {
			return ListPage{}, fmt.Errorf("scan listing: %w", err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		return ListPage{}, fmt.Errorf("list items: %w", err)
	}

	hasMore := len(got) > p.Limit
	if hasMore {
		got = got[:p.Limit]
	}
	if backwards {
		for i, j := 0, len(got)-1; i < j; i, j = i+1, j-1 {
			got[i], got[j] = got[j], got[i]
		}
	}

	page := ListPage{Total: total}
	for _, r := range got {
		page.Items = append(page.Items, r.ListedItem)
	}
	if len(got) == 0 {
		return page, nil
	}
	mk := func(r row, d string) string {
		return codec.Encode(Cursor{S: p.Sort, V: r.sortValue, ID: r.ID, D: d})
	}
	first, last := got[0], got[len(got)-1]
	if backwards || hasMore {
		page.NextCursor = mk(last, DirNext)
	}
	if (backwards && hasMore) || (!backwards && cur != nil) {
		page.PrevCursor = mk(first, DirPrev)
	}
	return page, nil
}

// ParseListQuery reads the listing filters of typ through get (a query
// lookup); malformed numbers are ErrInvalidQuery. Limit stays 0 (default).
func ParseListQuery(get func(string) string, typ string) (ListParams, error) {
	p := ListParams{
		Type: typ, Genre: get("genre"), Platform: get("platform"),
		Sort: get("sort"), Cursor: get("cursor"),
	}

	if v := get("year"); v != "" {
		year, err := strconv.Atoi(v)
		if err != nil {
			return p, ErrInvalidQuery
		}
		p.Year = year
	}
	stars := func(name string) (*float64, error) {
		v := get(name)
		if v == "" {
			return nil, nil
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, ErrInvalidQuery
		}

		return &f, nil
	}

	var err error
	if p.MinScore, err = stars("min_score"); err != nil {
		return p, err
	}
	if p.MaxScore, err = stars("max_score"); err != nil {
		return p, err
	}
	for _, key := range CategoryKeys[typ] {
		v, err := stars("min_" + key)
		if err != nil {
			return p, err
		}
		if v != nil {
			if p.MinCategory == nil {
				p.MinCategory = map[string]float64{}
			}
			p.MinCategory[key] = *v
		}
	}

	return p, nil
}
