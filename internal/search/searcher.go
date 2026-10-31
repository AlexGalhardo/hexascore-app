package search

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// Audience values mirror catalog.Audience (SDD 01 §6.1); kept as strings so this package stays free of app imports.
const (
	AudienceSite = "site"
	AudienceAPI  = "api"
)

const (
	maxReachable = 200
	maxPage      = 10
	suggestFetch = 24
	suggestCap   = 8
	suggestItems = 6
	suggestEnts  = 2
	searchTO     = 800 * time.Millisecond
	suggestTO    = 150 * time.Millisecond
	p99TTL       = 24 * time.Hour
)

var (
	itemTypes     = []string{"game", "movie", "series", "book"}
	entityKinds   = []string{"person", "organization", "franchise"}
	searchLocale  = map[string]bool{"": true, "en": true, "pt-BR": true, "es": true}
	suggestEntset = map[string]bool{"person": true, "organization": true, "franchise": true}
)

// Searcher is the single search entry point; an engine swap stays behind it (SDD 07 §6, D-16).
type Searcher interface {
	Search(ctx context.Context, q Query) (Results, error)
	Suggest(ctx context.Context, q SuggestQuery) ([]Suggestion, error)
	SearchEntities(ctx context.Context, q EntityQuery) (EntityResults, error)
}

// Query is an item search. Audience "api" ranks only allowlisted-source text; "site" ranks everything (07-7).
type Query struct {
	Text     string
	Locale   string   // "en" | "pt-BR" | "es" | "" = all
	Types    []string // empty = all
	Page     int      // 1..10, 0 = 1
	PerPage  int      // 1..50, 0 = 20
	Audience string   // "site" | "api", required
}

// EditionMatch names the edition that made a book work match (D-10).
type EditionMatch struct {
	ID, Title, ISBN13 string
}

// Hit is one ranked item. Title is the display title (UI locale, fallback en); MatchedTitle is set when the
// title that matched differs from it.
type Hit struct {
	ItemID, Type, Slug, Title string
	MatchedTitle              string
	MatchedEdition            *EditionMatch
	Rank                      float64
}

// Group is one type's block of the grouped site page.
type Group struct {
	Type  string
	Hits  []Hit
	Total int
}

// Results holds a flat page in Hits, or, for the site audience without a type filter, up to 8 hits per type in
// Groups. DidYouMean is a suggest label for zero-result site searches, never for the API (its labels are unfiltered).
type Results struct {
	Hits       []Hit
	Groups     []Group
	Total      int
	TooShort   bool
	DidYouMean *string
}

// SuggestQuery is an autocomplete lookup; Kinds filters search_suggest.kind.
type SuggestQuery struct {
	Text   string
	Locale string
	Kinds  []string
}

// Suggestion is one autocomplete entry; the controller builds the path and thumbnail URL.
type Suggestion struct {
	RefID, Kind, Title, Slug, ThumbKey string
	Year                               int
}

// EntityQuery searches person, organization and franchise entities.
type EntityQuery struct {
	Text     string
	Locale   string
	Kinds    []string // empty = all three
	Page     int
	PerPage  int
	Audience string
}

// EntityHit is one ranked entity.
type EntityHit struct {
	ID, Kind, Slug, Name string
	ItemCount            int
	Rank                 float64
}

// EntityResults is one page of entities.
type EntityResults struct {
	Hits  []EntityHit
	Total int
}

// Postgres implements Searcher on pg_trgm and the hs_* text search configurations.
type Postgres struct {
	db         *database.DB
	apiSources []string

	mu      sync.Mutex
	p99     float64
	p99From time.Time
}

// New builds the searcher; apiSources is the CATALOG_API_SOURCES allowlist used for the "api" audience.
func New(db *database.DB, apiSources []string) *Postgres {
	return &Postgres{db: db, apiSources: apiSources}
}

var _ Searcher = (*Postgres)(nil)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidQuery, fmt.Sprintf(format, args...))
}

func validatePaging(audience string, page, perPage *int) error {
	if audience != AudienceSite && audience != AudienceAPI {
		return invalid("unknown audience %q", audience)
	}

	if *page == 0 {
		*page = 1
	}

	if *perPage == 0 {
		*perPage = 20
	}

	switch {
	case *page < 1 || *page > maxPage:
		return invalid("page must be 1..%d", maxPage)
	case *perPage < 1 || *perPage > 50:
		return invalid("per page must be 1..50")
	case *page**perPage > maxReachable:
		return invalid("only the first %d results are reachable", maxReachable)
	}

	return nil
}

func validateList(values, allowed []string, what string) error {
	for _, v := range values {
		if !contains(allowed, v) {
			return invalid("unknown %s %q", what, v)
		}
	}

	return nil
}

func contains(list []string, v string) bool { return slices.Contains(list, v) }

// textArray renders a Postgres text[] literal; nil renders NULL (as a nil any).
func textArray(values []string) any {
	if values == nil {
		return nil
	}

	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v) + `"`
	}

	return "{" + strings.Join(parts, ",") + "}"
}

// sources returns the allowlist argument: NULL for the site, a (possibly empty, so deny-all) array for the API.
func (s *Postgres) sources(audience string) any {
	if audience == AudienceSite {
		return nil
	}

	return textArray(append([]string{}, s.apiSources...))
}

func isTimeout(err error) bool {
	return err != nil &&
		(errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "statement timeout"))
}

// withSettings runs fn in a transaction with the per-transaction trigram threshold and statement timeout
// (SDD 07 NFR-07-3); set_config(..., true) is SET LOCAL.
func (s *Postgres) withSettings(
	ctx context.Context, timeout time.Duration, guc, threshold string, fn func(tx *database.Tx) error,
) error {
	return s.db.Tx(ctx, func(tx *database.Tx) error {
		// enable_seqscan off: the planner prices the trigram operators (<%, %) as cheap functions, so on a catalogue
		// of production size it prefers a sequential scan of item_alt_titles over its GIN index; word_similarity
		// then runs on every row (about 160 ms per 50k rows, over the 800 ms budget on the full table). The text
		// tables are only reached through their indexes, so the penalty never hides a better plan.
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config($1, $2, true), set_config('statement_timeout', $3, true),
			        set_config('enable_seqscan', 'off', true)`,
			guc, threshold, strconv.FormatInt(timeout.Milliseconds(), 10)); err != nil {
			return fmt.Errorf("set search settings: %w", err)
		}

		return fn(tx)
	})
}

func (s *Postgres) popularityScale(ctx context.Context) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if time.Since(s.p99From) < p99TTL && !s.p99From.IsZero() {
		return s.p99
	}

	var v sql.NullFloat64

	err := s.db.QueryRowContext(ctx,
		`SELECT percentile_cont(0.99) WITHIN GROUP (ORDER BY popularity_prior) FROM items WHERE visibility = 'public'`,
	).Scan(&v)
	if err != nil {
		return s.p99 // keep the last value; 0 makes the popularity term vanish
	}

	s.p99, s.p99From = v.Float64, time.Now()

	return s.p99
}

// Search ranks items by the §7.3 tiers.
func (s *Postgres) Search(ctx context.Context, q Query) (Results, error) {
	if err := validatePaging(q.Audience, &q.Page, &q.PerPage); err != nil {
		return Results{}, err
	}

	if err := validateList(q.Types, itemTypes, "type"); err != nil {
		return Results{}, err
	}

	if !searchLocale[q.Locale] {
		return Results{}, invalid("unknown locale %q", q.Locale)
	}

	p, err := Parse(q.Text)
	if err != nil {
		return Results{}, err
	}

	if p.Empty || p.TooShort || p.ExclusionOnly() {
		return Results{TooShort: p.TooShort}, nil
	}

	var types any
	if len(q.Types) > 0 {
		types = textArray(q.Types)
	}

	var isbn any
	if p.ISBN13 != "" {
		isbn = p.ISBN13
	}

	grouped := len(q.Types) == 0 && q.Audience == AudienceSite
	limit, offset := q.PerPage, (q.Page-1)*q.PerPage

	var res Results

	scale := s.popularityScale(ctx)

	run := func(tx *database.Tx) error {
		sqlText := itemSQL(q.Locale, q.Audience == AudienceAPI, p.ISBN13 != "", grouped, p.Excluded != "")

		rows, err := tx.QueryContext(ctx, sqlText, p.Text, p.Match, types, scale, q.Locale,
			s.sources(q.Audience), limit, offset, isbn, p.Excluded)
		if err != nil {
			return fmt.Errorf("search items: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				h              Hit
				mt, k, eid, in sql.NullString
				total          int
			)

			if err := rows.Scan(
				&h.ItemID,
				&h.Type,
				&h.Slug,
				&h.Title,
				&mt,
				&k,
				&eid,
				&in,
				&h.Rank,
				&total,
			); err != nil {
				return err
			}

			if mt.Valid && mt.String != h.Title {
				h.MatchedTitle = mt.String
			}

			if k.String == "e" && eid.Valid {
				h.MatchedEdition = &EditionMatch{ID: eid.String, Title: mt.String, ISBN13: in.String}
			}

			if grouped {
				res.addToGroup(h, total)
			} else {
				res.Hits = append(res.Hits, h)
				res.Total = total
			}
		}

		return rows.Err()
	}

	if err := s.withSettings(ctx, searchTO, "pg_trgm.word_similarity_threshold", "0.2", run); err != nil {
		if isTimeout(err) {
			return Results{}, ErrTimeout
		}

		return Results{}, err
	}

	if grouped {
		res.orderGroups()
	}

	if res.empty() && q.Audience == AudienceSite && len([]rune(p.Match)) >= 3 {
		res.DidYouMean = s.didYouMean(ctx, p.Match)
	}

	return res, nil
}

func (r *Results) addToGroup(h Hit, total int) {
	for i := range r.Groups {
		if r.Groups[i].Type == h.Type {
			r.Groups[i].Hits = append(r.Groups[i].Hits, h)

			return
		}
	}

	r.Groups = append(r.Groups, Group{Type: h.Type, Hits: []Hit{h}, Total: total})
	r.Total += total
}

func (r *Results) orderGroups() {
	ordered := make([]Group, 0, len(r.Groups))

	for _, t := range itemTypes {
		for _, g := range r.Groups {
			if g.Type == t {
				ordered = append(ordered, g)
			}
		}
	}

	r.Groups = ordered
}

func (r *Results) empty() bool { return len(r.Hits) == 0 && len(r.Groups) == 0 }

// didYouMean returns the closest suggest label (similarity >= 0.3, popularity tie-break) or nil.
func (s *Postgres) didYouMean(ctx context.Context, text string) *string {
	var label *string

	_ = s.withSettings(ctx, suggestTO, "pg_trgm.similarity_threshold", "0.3", func(tx *database.Tx) error {
		var l string

		err := tx.QueryRowContext(ctx,
			`SELECT label FROM search_suggest WHERE label_norm % hs_norm($1)
			 ORDER BY similarity(label_norm, hs_norm($1)) DESC, popularity DESC, label LIMIT 1`, text).Scan(&l)
		if err == nil {
			label = &l
		}

		return nil
	})

	return label
}

// Suggest returns up to 8 autocomplete entries (6 items, 2 entities): prefix first, trigram when under 8 refs.
// A timeout yields an empty list, not an error (SDD 07 §8).
func (s *Postgres) Suggest(ctx context.Context, q SuggestQuery) ([]Suggestion, error) {
	if !searchLocale[q.Locale] {
		return nil, invalid("unknown locale %q", q.Locale)
	}

	if err := validateList(q.Kinds, append(append([]string{}, itemTypes...), entityKinds...), "kind"); err != nil {
		return nil, err
	}

	p, err := Parse(q.Text)
	if err != nil {
		return nil, err
	}

	if p.Empty || p.TooShort || p.Match == "" {
		return []Suggestion{}, nil
	}

	var kinds any
	if len(q.Kinds) > 0 {
		kinds = textArray(q.Kinds)
	}

	const cols = `ref_id::text, kind, locale, label, slug, coalesce(year, 0), coalesce(thumb_key, '')`

	const filter = `AND ($2::text = '' OR locale IN ($2::text, '*')) AND ($3::text[] IS NULL OR kind = ANY($3))`

	var out []Suggestion

	err = s.withSettings(ctx, suggestTO, "pg_trgm.similarity_threshold", "0.3", func(tx *database.Tx) error {
		var all []suggestRow

		prefix, err := scanSuggest(ctx, tx, `SELECT `+cols+` FROM search_suggest
			WHERE label_norm ^@ hs_norm($1) `+filter+` ORDER BY popularity DESC, label LIMIT `+strconv.Itoa(suggestFetch),
			p.Match, q.Locale, kinds)
		if err != nil {
			return err
		}

		all = append(all, prefix...)

		if distinctRefs(all) < suggestCap && len([]rune(p.Match)) >= 3 {
			fuzzy, err := scanSuggest(ctx, tx, `SELECT `+cols+` FROM search_suggest
				WHERE label_norm % hs_norm($1) `+filter+` ORDER BY similarity(label_norm, hs_norm($1)) DESC, popularity DESC, label
				LIMIT `+strconv.Itoa(suggestFetch), p.Match, q.Locale, kinds)
			if err != nil {
				return err
			}

			all = append(all, fuzzy...)
		}

		out = dedupeSuggest(all, q.Locale)

		return nil
	})
	if isTimeout(err) {
		return []Suggestion{}, nil
	}

	return out, err
}

type suggestRow struct {
	Suggestion
	locale string
}

func scanSuggest(ctx context.Context, tx *database.Tx, query string, args ...any) ([]suggestRow, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("suggest: %w", err)
	}
	defer rows.Close()

	var out []suggestRow

	for rows.Next() {
		var r suggestRow
		if err := rows.Scan(&r.RefID, &r.Kind, &r.locale, &r.Title, &r.Slug, &r.Year, &r.ThumbKey); err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, rows.Err()
}

func distinctRefs(rows []suggestRow) int {
	seen := map[string]bool{}
	for _, r := range rows {
		seen[r.RefID] = true
	}

	return len(seen)
}

// dedupeSuggest keeps one entry per ref (the UI-locale label wins over a '*' alternative one), in first-seen order,
// capped at 6 items and 2 entities.
func dedupeSuggest(rows []suggestRow, ui string) []Suggestion {
	best := map[string]suggestRow{}
	var order []string

	for _, r := range rows {
		cur, ok := best[r.RefID]
		if !ok {
			order = append(order, r.RefID)
			best[r.RefID] = r

			continue
		}

		if cur.locale != ui && r.locale == ui {
			best[r.RefID] = r
		}
	}

	out := []Suggestion{}
	items, ents := 0, 0

	for _, id := range order {
		r := best[id]

		if suggestEntset[r.Kind] {
			if ents == suggestEnts {
				continue
			}

			ents++
		} else {
			if items == suggestItems {
				continue
			}

			items++
		}

		out = append(out, r.Suggestion)
	}

	return out
}

// SearchEntities ranks people, organisations and franchises by name (SDD 07 §7.4).
func (s *Postgres) SearchEntities(ctx context.Context, q EntityQuery) (EntityResults, error) {
	if err := validatePaging(q.Audience, &q.Page, &q.PerPage); err != nil {
		return EntityResults{}, err
	}

	if err := validateList(q.Kinds, entityKinds, "kind"); err != nil {
		return EntityResults{}, err
	}

	if !searchLocale[q.Locale] {
		return EntityResults{}, invalid("unknown locale %q", q.Locale)
	}

	p, err := Parse(q.Text)
	if err != nil {
		return EntityResults{}, err
	}

	if p.Empty || p.TooShort || p.Match == "" {
		return EntityResults{}, nil
	}

	kinds := q.Kinds
	if len(kinds) == 0 {
		kinds = entityKinds
	}

	var res EntityResults

	err = s.withSettings(ctx, searchTO, "pg_trgm.word_similarity_threshold", "0.2", func(tx *database.Tx) error {
		rows, err := tx.QueryContext(ctx, entitySQL(q.Audience == AudienceAPI), p.Match, textArray(kinds), q.Locale,
			s.sources(q.Audience), q.PerPage, (q.Page-1)*q.PerPage)
		if err != nil {
			return fmt.Errorf("search entities: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var h EntityHit
			if err := rows.Scan(&h.ID, &h.Kind, &h.Slug, &h.Name, &h.Rank, &h.ItemCount, &res.Total); err != nil {
				return err
			}

			res.Hits = append(res.Hits, h)
		}

		return rows.Err()
	})
	if isTimeout(err) {
		return EntityResults{}, ErrTimeout
	}

	return res, err
}
