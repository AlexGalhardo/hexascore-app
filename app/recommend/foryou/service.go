package foryou

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
)

const (
	// StaleAfter is the age after which a stored list is served stale while a refresh is queued.
	StaleAfter = 24 * time.Hour
	// StaleEvents is the number of newer rating_events that makes a stored list stale.
	StaleEvents = 5

	candidatesPerFeature = 200
	topSimilarItems      = 30
	topFeatures          = 20
)

// RecommendUser is the queued recommend_user job (SDD 08 section 6 "Jobs"). Register it with Retries(2), Timeout(30s).
type RecommendUser struct {
	UserID int64  `json:"user_id"`
	Type   string `json:"type"`
}

// Kind names the job on the queue.
func (RecommendUser) Kind() string { return "recommend.user" }

// ReasonView is the structured explanation of a pick; the client words it through the i18n catalogues.
type ReasonView struct {
	Kind   string   `json:"kind"`
	ItemID string   `json:"itemId"`
	Title  string   `json:"title"`
	Stars  *float64 `json:"stars,omitempty"`
}

// Item is one served recommendation.
type Item struct {
	Item   catalog.Highlight `json:"item"`
	Reason *ReasonView       `json:"reason"`
}

// List is what "For you" serves. Basis is BasisOnboarding when the user has too few positive items; Items are then
// the trending of the type, without reasons.
type List struct {
	Basis string `json:"basis"`
	Data  []Item `json:"data"`
}

// Service computes, stores and serves "For you". Enqueue may be nil, which disables the background refresh.
type Service struct {
	DB      *database.DB
	Now     func() time.Time
	Enqueue func(ctx context.Context, j RecommendUser) error
}

// OptedOut reports users.personalization_opt_out.
func (s *Service) OptedOut(ctx context.Context, userID int64) (bool, error) {
	var out bool

	err := s.DB.QueryRowContext(ctx, `SELECT personalization_opt_out FROM users WHERE id = $1`, userID).Scan(&out)

	return out, err
}

// Get serves the list of the user for one type. An opted-out user gets no list and no stored rows (D-29). A stored
// row older than StaleAfter, or with StaleEvents newer rating_events, is served as is while a refresh is queued; a
// user with no row waits for a synchronous compute.
func (s *Service) Get(ctx context.Context, userID int64, typ, locale string) (List, error) {
	if out, err := s.OptedOut(ctx, userID); err != nil || out {
		if err == nil {
			err = s.Purge(ctx, userID)
		}

		return List{Basis: BasisOnboarding, Data: []Item{}}, err
	}

	ids, reasons, at, events, found, err := s.stored(ctx, userID, typ)
	if err != nil {
		return List{}, err
	}

	if !found {
		if ids, reasons, err = s.Refresh(ctx, userID, typ); err != nil {
			return List{}, err
		}
	} else if s.stale(ctx, userID, at, events) && s.Enqueue != nil {
		if err := s.Enqueue(ctx, RecommendUser{UserID: userID, Type: typ}); err != nil {
			return List{}, err
		}
	}

	return s.render(ctx, userID, typ, locale, ids, reasons)
}

// Purge deletes the stored rows of the user.
func (s *Service) Purge(ctx context.Context, userID int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM user_recommendations WHERE user_id = $1`, userID)

	return err
}

func (s *Service) stored(ctx context.Context, userID int64, typ string) (
	ids, reasons []string, at time.Time, events int64, found bool, err error,
) {
	var a, b string

	err = s.DB.QueryRowContext(ctx, `SELECT item_ids::text, reason_item_ids::text, computed_at, events_at_compute
FROM user_recommendations WHERE user_id = $1 AND item_type = $2`, userID, typ).Scan(&a, &b, &at, &events)
	if err == sql.ErrNoRows {
		return nil, nil, at, 0, false, nil
	}

	return splitArray(a), splitArray(b), at, events, err == nil, err
}

func (s *Service) stale(ctx context.Context, userID int64, at time.Time, events int64) bool {
	if s.Now().Sub(at) > StaleAfter {
		return true
	}

	var n int64
	if err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM rating_events WHERE user_id = $1`, userID).
		Scan(&n); err != nil {
		return false
	}

	return n-events >= StaleEvents
}

// Refresh recomputes and stores the list of the user. It writes nothing for an opted-out user.
func (s *Service) Refresh(ctx context.Context, userID int64, typ string) (ids, reasons []string, err error) {
	if out, err := s.OptedOut(ctx, userID); err != nil || out {
		return nil, nil, err
	}

	now := s.Now().UTC()

	var events int64
	if err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM rating_events WHERE user_id = $1`, userID).
		Scan(&events); err != nil {
		return nil, nil, err
	}

	res, err := s.compute(ctx, userID, typ, now)
	if err != nil {
		return nil, nil, err
	}

	ids, reasons = []string{}, []string{}
	for _, p := range res.Items {
		ids, reasons = append(ids, p.ID), append(reasons, p.ReasonID)
	}

	_, err = s.DB.ExecContext(ctx, `INSERT INTO user_recommendations
(user_id, item_type, item_ids, reason_item_ids, events_at_compute, computed_at)
VALUES ($1, $2, $3::uuid[], $4::uuid[], $5, $6)
ON CONFLICT (user_id, item_type) DO UPDATE SET item_ids = EXCLUDED.item_ids,
  reason_item_ids = EXCLUDED.reason_item_ids, events_at_compute = EXCLUDED.events_at_compute,
  computed_at = EXCLUDED.computed_at`, userID, typ, pgArray(ids), pgArray(reasons), events, now)

	return ids, reasons, err
}

// Run is the recommend_user job body: it exits without writing when the user opted out.
func (s *Service) Run(ctx context.Context, j RecommendUser) error {
	_, _, err := s.Refresh(ctx, j.UserID, j.Type)

	return err
}

func (s *Service) render(ctx context.Context, userID int64, typ, locale string, ids, reasons []string) (List, error) {
	if len(ids) == 0 {
		return s.onboarding(ctx, typ, locale)
	}

	cards, err := catalog.HighlightsByID(ctx, s.DB, ids, locale)
	if err != nil {
		return List{}, err
	}

	why, err := s.reasons(ctx, userID, locale, reasons)
	if err != nil {
		return List{}, err
	}

	reasonOf := make(map[string]string, len(ids))
	for i, id := range ids {
		if i < len(reasons) {
			reasonOf[id] = reasons[i]
		}
	}

	out := List{Basis: BasisPersonal, Data: make([]Item, 0, len(cards))}
	for _, c := range cards {
		out.Data = append(out.Data, Item{Item: c, Reason: why[reasonOf[c.ID]]})
	}

	return out, nil
}

func (s *Service) onboarding(ctx context.Context, typ, locale string) (List, error) {
	out := List{Basis: BasisOnboarding, Data: []Item{}}

	_, rows, err := trending.Latest(ctx, s.DB, typ)
	if err != nil {
		return out, err
	}

	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ItemID
	}

	cards, err := catalog.HighlightsByID(ctx, s.DB, ids, locale)
	for _, c := range cards {
		out.Data = append(out.Data, Item{Item: c})
	}

	return out, err
}

// reasons words each reason item from the user's current rating or status; an item the user no longer rated or
// statused has no reason.
func (s *Service) reasons(
	ctx context.Context,
	userID int64,
	locale string,
	ids []string,
) (map[string]*ReasonView, error) {
	out := map[string]*ReasonView{}
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := s.DB.QueryContext(ctx, `SELECT i.id::text, coalesce(t.title, te.title, i.original_title, i.slug),
  r.overall::float8, s.status::text
FROM items i
LEFT JOIN ratings r ON r.item_id = i.id AND r.user_id = $2
LEFT JOIN item_status s ON s.item_id = i.id AND s.user_id = $2
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $3
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE i.id = ANY($1::uuid[])`, pgArray(ids), userID, locale)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			id, title string
			overall   sql.NullFloat64
			status    sql.NullString
		)
		if err := rows.Scan(&id, &title, &overall, &status); err != nil {
			return nil, err
		}

		v := &ReasonView{ItemID: id, Title: title}

		switch {
		case overall.Valid:
			stars := overall.Float64 / 20
			v.Kind, v.Stars = "rated", &stars
		case status.String == "completed":
			v.Kind = "completed"
		case status.String == "want":
			v.Kind = "wanted"
		default:
			continue
		}

		out[id] = v
	}

	return out, rows.Err()
}

// weights memoises g_f for one type: N_t once, n_f per feature on first use.
type weights struct {
	db    *database.DB
	typ   string
	total int
	n     map[similar.Feature]int
}

func newWeights(ctx context.Context, db *database.DB, typ string) (*weights, error) {
	w := &weights{db: db, typ: typ, n: map[similar.Feature]int{}}

	err := db.QueryRowContext(ctx, `SELECT count(*) FROM items WHERE type = $1 AND visibility = 'public'`, typ).
		Scan(&w.total)

	return w, err
}

func (w *weights) ensure(ctx context.Context, sets ...map[similar.Feature]float64) error {
	var roles, ents []string

	for _, set := range sets {
		for f := range set {
			if _, ok := w.n[f]; !ok {
				w.n[f] = 0
				roles, ents = append(roles, f.Role), append(ents, f.Entity)
			}
		}
	}

	if len(roles) == 0 {
		return nil
	}

	rows, err := w.db.QueryContext(ctx, `SELECT ie.role, ie.entity_id::text, count(DISTINCT ie.item_id)
FROM item_entities ie JOIN items i ON i.id = ie.item_id AND i.type = $1 AND i.visibility = 'public'
WHERE (ie.role, ie.entity_id::text) IN (SELECT * FROM unnest($2::text[], $3::text[]))
GROUP BY 1, 2`, w.typ, pgArray(roles), pgArray(ents))
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			f similar.Feature
			n int
		)
		if err := rows.Scan(&f.Role, &f.Entity, &n); err != nil {
			return err
		}

		w.n[f] = n
	}

	return rows.Err()
}

func (w *weights) g(f similar.Feature) float64 { return similar.Weight(f.Role, w.total, w.n[f]) }

func (s *Service) compute(ctx context.Context, userID int64, typ string, now time.Time) (Result, error) {
	parents, err := similar.LoadParents(ctx, s.DB)
	if err != nil {
		return Result{}, err
	}

	ints, err := s.interactions(ctx, userID, typ, parents)
	if err != nil {
		return Result{}, err
	}

	w, err := newWeights(ctx, s.DB, typ)
	if err != nil {
		return Result{}, err
	}

	for _, in := range ints {
		if err := w.ensure(ctx, in.Features); err != nil {
			return Result{}, err
		}
	}

	p := BuildProfile(ints, now, w.g)
	if p.Positives < MinPositives {
		return Result{Basis: BasisOnboarding}, nil
	}

	cands, err := s.candidates(ctx, typ, ints, p, now, parents, w)
	if err != nil {
		return Result{}, err
	}

	return Recommend(ints, cands, now, w.g), nil
}

func (s *Service) interactions(
	ctx context.Context,
	userID int64,
	typ string,
	parents map[string][]string,
) ([]Interaction, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT i.id::text, coalesce(te.title, i.original_title, i.slug),
  r.overall::float8, s.status::text, coalesce(r.updated_at, s.updated_at)
FROM items i
LEFT JOIN ratings r ON r.item_id = i.id AND r.user_id = $1
LEFT JOIN item_status s ON s.item_id = i.id AND s.user_id = $1
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE i.type = $2 AND (r.user_id IS NOT NULL OR s.user_id IS NOT NULL)`, userID, typ)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var (
		ints []Interaction
		ids  []string
	)

	for rows.Next() {
		var (
			in      Interaction
			overall sql.NullFloat64
			status  sql.NullString
		)
		if err := rows.Scan(&in.ItemID, &in.Title, &overall, &status, &in.At); err != nil {
			return nil, err
		}

		if overall.Valid {
			in.Overall = &overall.Float64
		}

		in.Status = status.String
		ints, ids = append(ints, in), append(ids, in.ItemID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	entries, err := similar.LoadEntries(ctx, s.DB, typ, ids)
	if err != nil {
		return nil, err
	}

	for i := range ints {
		ints[i].Features = similar.ExtractFeatures(entries[ints[i].ItemID], parents)
	}

	return ints, nil
}

// candidates are the item_similar neighbours of the user's top items plus the most popular items of the user's
// strongest features (SDD 08 section 7.5), with the facts Recommend needs.
func (s *Service) candidates(ctx context.Context, typ string, ints []Interaction, p Profile, now time.Time,
	parents map[string][]string, w *weights,
) ([]Candidate, error) {
	seen := map[string]bool{}
	for _, in := range ints {
		seen[in.ItemID] = true
	}

	pool := map[string]bool{}
	add := func(q string, args ...any) error {
		rows, err := s.DB.QueryContext(ctx, q, args...)
		if err != nil {
			return err
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}

			if !seen[id] {
				pool[id] = true
			}
		}

		return rows.Err()
	}

	top := TopItems(ints, now, topSimilarItems)
	if err := add(
		`SELECT unnest(similar_ids)::text FROM item_similar WHERE item_id = ANY($1::uuid[])`,
		pgArray(top),
	); err != nil {
		return nil, err
	}

	for _, f := range TopFeatures(p, topFeatures) {
		if err := add(`SELECT ie.item_id::text FROM item_entities ie
JOIN items i ON i.id = ie.item_id AND i.type = $1 AND i.visibility = 'public'
WHERE ie.role = $2 AND ie.entity_id = $3::uuid ORDER BY i.popularity_prior DESC, i.id LIMIT `+strconv.Itoa(candidatesPerFeature),
			typ, f.Role, f.Entity); err != nil {
			return nil, err
		}
	}

	ids := make([]string, 0, len(pool))
	for id := range pool {
		ids = append(ids, id)
	}

	entries, err := similar.LoadEntries(ctx, s.DB, typ, ids)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.QueryContext(ctx, `SELECT i.id::text, coalesce(te.title, i.original_title, i.slug),
  i.release_status = 'released' AND (i.release_date IS NULL OR i.release_date <= $2::date),
  CASE WHEN i.rating_count >= 5 THEN i.score::float8 END
FROM items i LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE i.id = ANY($1::uuid[]) AND i.type = $3 AND i.visibility = 'public'`, pgArray(ids), now, typ)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []Candidate

	for rows.Next() {
		c := Candidate{Public: true}

		var bayes sql.NullFloat64
		if err := rows.Scan(&c.ID, &c.Title, &c.Released, &bayes); err != nil {
			return nil, err
		}

		if bayes.Valid {
			c.Bayes = &bayes.Float64
		}

		for _, e := range entries[c.ID] {
			if e.Role == "franchise" {
				c.Franchise = e.Entity
			}

			if e.Role == "genre" && e.Position == 0 {
				c.PrimaryGenre = e.Entity
			}
		}

		c.Features = similar.ExtractFeatures(entries[c.ID], parents)
		out = append(out, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, c := range out {
		if err := w.ensure(ctx, c.Features); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func pgArray(vals []string) string { return "{" + strings.Join(vals, ",") + "}" }

func splitArray(s string) []string {
	s = strings.Trim(s, "{}")
	if s == "" {
		return []string{}
	}

	return strings.Split(s, ",")
}
