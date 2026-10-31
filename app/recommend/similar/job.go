package similar

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

const (
	// ChunkSize is the number of subject items read and written per transaction.
	ChunkSize = 500
	// CandidateShare is the largest share of N_t a feature may cover and still generate candidates (5 %).
	CandidateShare = 0.05
	// DefaultPause is the sleep between chunks.
	DefaultPause = 50 * time.Millisecond
)

// Types are the catalogue types similarity is computed for; neighbours are always of the subject's type.
var Types = []string{"game", "movie", "series", "book"}

// Options tunes a run. The zero value is the site variant: every relation, DefaultPause between chunks.
type Options struct {
	// Sources, when non-empty, selects the API variant: only item_entities rows whose source is listed count
	// (CATALOG_API_SOURCES, SDD 08 open issue 08-9).
	Sources []string
	// Pause between chunks; zero means DefaultPause, negative disables it.
	Pause time.Duration
	// Now stamps computed_at.
	Now time.Time
}

// Compute refreshes item_similar for every public item and returns the number of rows written. An item whose feature
// hash is unchanged since the last run is skipped (SDD 08 section 7.4).
// shortcut: n_f counts items holding the feature directly (a franchise parent is not counted through its children), and
// candidates come from direct holders of any feature in F_i including parents; revisit with the 500k benchmark (08-11).
func Compute(ctx context.Context, db *database.DB, opt Options) (int, error) {
	if opt.Pause == 0 {
		opt.Pause = DefaultPause
	}

	parents, err := loadParents(ctx, db)
	if err != nil {
		return 0, err
	}

	written := 0

	for _, typ := range Types {
		n, err := computeType(ctx, db, typ, parents, opt)
		written += n

		if err != nil {
			return written, fmt.Errorf("similar %s: %w", typ, err)
		}
	}

	return written, nil
}

// sourceFilter is the item_entities predicate shared by every read; the cast cut keeps n_f consistent with F_i.
func sourceFilter(opt Options, alias string, argPos int) (string, []any) {
	where := fmt.Sprintf("(%[1]s.role <> 'cast' OR %[1]s.position < %d)", alias, CastPositions)
	if len(opt.Sources) == 0 {
		return where, nil
	}

	return where + fmt.Sprintf(" AND %s.source = ANY($%d::text[])", alias, argPos), []any{pgArray(opt.Sources)}
}

func computeType(
	ctx context.Context,
	db *database.DB,
	typ string,
	parents map[string][]string,
	opt Options,
) (int, error) {
	var total int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM items WHERE type = $1 AND visibility = 'public'`, typ).Scan(&total); err != nil {
		return 0, err
	}

	if total == 0 {
		return 0, nil
	}

	nf, err := loadCounts(ctx, db, typ, opt)
	if err != nil {
		return 0, err
	}

	weight := func(f Feature) float64 { return Weight(f.Role, total, nf[f]) }
	rare := func(f Feature) bool { return float64(nf[f]) <= CandidateShare*float64(total) }

	written, after := 0, ""

	for {
		ids, err := nextIDs(ctx, db, typ, after)
		if err != nil {
			return written, err
		}

		if len(ids) == 0 {
			return written, nil
		}

		after = ids[len(ids)-1]

		n, err := computeChunk(ctx, db, typ, ids, parents, weight, rare, opt)
		written += n

		if err != nil {
			return written, err
		}

		if err := pause(ctx, opt.Pause); err != nil {
			return written, err
		}
	}
}

func pause(ctx context.Context, d time.Duration) error {
	if d < 0 {
		return nil
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func nextIDs(ctx context.Context, db *database.DB, typ, after string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT id::text FROM items
WHERE type = $1 AND visibility = 'public' AND ($2 = '' OR id > $2::uuid) ORDER BY id LIMIT $3`, typ, after, ChunkSize)
	if err != nil {
		return nil, err
	}

	return scanStrings(rows)
}

func scanStrings(rows *sql.Rows) ([]string, error) {
	defer func() { _ = rows.Close() }()

	var out []string

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}

		out = append(out, s)
	}

	return out, rows.Err()
}

func loadParents(ctx context.Context, db *database.DB) (map[string][]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT from_id::text, to_id::text FROM entity_links WHERE relation = 'part_of'`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := map[string][]string{}

	for rows.Next() {
		var from, to string
		if err := rows.Scan(&from, &to); err != nil {
			return nil, err
		}

		out[from] = append(out[from], to)
	}

	return out, rows.Err()
}

func loadCounts(ctx context.Context, db *database.DB, typ string, opt Options) (map[Feature]int, error) {
	filter, args := sourceFilter(opt, "ie", 2)

	rows, err := db.QueryContext(ctx, `SELECT ie.role, ie.entity_id::text, count(DISTINCT ie.item_id)
FROM item_entities ie JOIN items i ON i.id = ie.item_id AND i.type = $1 AND i.visibility = 'public'
WHERE `+filter+` GROUP BY 1, 2`, append([]any{typ}, args...)...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := map[Feature]int{}

	for rows.Next() {
		var (
			f Feature
			n int
		)
		if err := rows.Scan(&f.Role, &f.Entity, &n); err != nil {
			return nil, err
		}

		out[f] = n
	}

	return out, rows.Err()
}

// loadEntries reads the item_entities rows and popularity of the given public items of typ.
func loadEntries(
	ctx context.Context,
	db *database.DB,
	typ string,
	ids []string,
	opt Options,
) (map[string][]Entry, map[string]float64, error) {
	filter, args := sourceFilter(opt, "ie", 3)

	rows, err := db.QueryContext(
		ctx,
		`SELECT i.id::text, i.popularity_prior, ie.role, coalesce(ie.entity_id::text, ''), coalesce(ie.position, 0)
FROM items i LEFT JOIN item_entities ie ON ie.item_id = i.id AND `+filter+`
WHERE i.id = ANY($1::uuid[]) AND i.type = $2 AND i.visibility = 'public'`,
		append([]any{pgArray(ids), typ}, args...)...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	entries := map[string][]Entry{}
	pop := map[string]float64{}

	for rows.Next() {
		var (
			id   string
			p    float64
			role sql.NullString
			ent  string
			pos  int
		)
		if err := rows.Scan(&id, &p, &role, &ent, &pos); err != nil {
			return nil, nil, err
		}

		pop[id] = p

		if role.Valid {
			entries[id] = append(entries[id], Entry{Role: role.String, Entity: ent, Position: pos})
		}
	}

	return entries, pop, rows.Err()
}

func computeChunk(ctx context.Context, db *database.DB, typ string, ids []string, parents map[string][]string,
	weight func(Feature) float64, rare func(Feature) bool, opt Options,
) (int, error) {
	entries, _, err := loadEntries(ctx, db, typ, ids, opt)
	if err != nil {
		return 0, err
	}

	old, err := currentHashes(ctx, db, ids)
	if err != nil {
		return 0, err
	}

	// Only items whose own feature hash changed pay for candidate generation (SDD 08 section 7.4).
	feats := make(map[string]map[Feature]float64, len(ids))
	hashes := make(map[string][]byte, len(ids))

	var (
		changed     []string
		roles, ents []string
	)

	seen := map[Feature]bool{}

	for _, id := range ids {
		f := ExtractFeatures(entries[id], parents)

		h := featureHash(f)
		if old[id] == hex.EncodeToString(h) {
			continue
		}

		feats[id], hashes[id] = f, h
		changed = append(changed, id)

		for ft := range f {
			if rare(ft) && !seen[ft] {
				seen[ft] = true
				roles = append(roles, ft.Role)
				ents = append(ents, ft.Entity)
			}
		}
	}

	if len(changed) == 0 {
		return 0, nil
	}

	byFeature, err := holders(ctx, db, typ, roles, ents, opt)
	if err != nil {
		return 0, err
	}

	cand := map[string][]string{}
	need := map[string]bool{}

	for _, id := range changed {
		set := map[string]bool{}

		for f := range feats[id] {
			if !rare(f) {
				continue
			}

			for _, o := range byFeature[f] {
				if o != id {
					set[o] = true
				}
			}
		}

		for o := range set {
			cand[id] = append(cand[id], o)
			need[o] = true
		}
	}

	candEntries, candPop, err := loadEntries(ctx, db, typ, keys(need), opt)
	if err != nil {
		return 0, err
	}

	type row struct {
		id   string
		hash []byte
		nb   []Scored
	}

	rows := make([]row, 0, len(changed))
	candFeats := map[string]map[Feature]float64{}

	for _, id := range changed {
		cs := make([]Candidate, 0, len(cand[id]))

		for _, o := range cand[id] {
			if _, public := candPop[o]; !public {
				continue
			}

			f, ok := candFeats[o]
			if !ok {
				f = ExtractFeatures(candEntries[o], parents)
				candFeats[o] = f
			}

			cs = append(cs, Candidate{ID: o, Features: f, Popularity: candPop[o]})
		}

		rows = append(rows, row{id: id, hash: hashes[id], nb: Filter(feats[id], cs, weight)})
	}

	subjects := make([][]Scored, len(rows))
	for i := range rows {
		subjects[i] = rows[i].nb
	}

	if err := fillCooccur(ctx, db, changed, subjects); err != nil {
		return 0, err
	}

	for i := range rows {
		rows[i].nb = Order(rows[i].nb)
	}

	// One multi-row statement per chunk: a round trip per row dominated the run time.
	var (
		sb   strings.Builder
		args = make([]any, 0, len(rows)*6)
	)

	for i, r := range rows {
		var nbIDs, sim, co []string

		for _, s := range r.nb {
			nbIDs = append(nbIDs, s.ID)
			sim = append(sim, fmt.Sprint(s.Sim))
			co = append(co, strconv.Itoa(s.Cooccur))
		}

		if i > 0 {
			sb.WriteString(",")
		}

		n := i * 6
		fmt.Fprintf(&sb, "($%d, $%d::uuid[], $%d::real[], $%d::int[], $%d, $%d)", n+1, n+2, n+3, n+4, n+5, n+6)
		args = append(args, r.id, pgArray(nbIDs), pgArray(sim), pgArray(co), r.hash, opt.Now)
	}

	err = db.Tx(ctx, func(tx *database.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO item_similar (item_id, similar_ids, scores, cooccur, feature_hash, computed_at)
VALUES `+sb.String()+`
ON CONFLICT (item_id) DO UPDATE SET similar_ids = EXCLUDED.similar_ids, scores = EXCLUDED.scores,
  cooccur = EXCLUDED.cooccur, feature_hash = EXCLUDED.feature_hash, computed_at = EXCLUDED.computed_at`, args...)

		return err
	})

	return len(rows), err
}

func holders(
	ctx context.Context,
	db *database.DB,
	typ string,
	roles, ents []string,
	opt Options,
) (map[Feature][]string, error) {
	out := map[Feature][]string{}
	if len(roles) == 0 {
		return out, nil
	}

	filter, args := sourceFilter(opt, "ie", 3)

	// No join to items: with fresh bulk loads the planner has no statistics and picks a quadratic nested loop. Visibility
	// and type are enforced when the candidates are loaded.
	rows, err := db.QueryContext(ctx, `SELECT f.role, f.entity_id::text, ie.item_id::text
FROM unnest($1::text[], $2::uuid[]) AS f(role, entity_id)
CROSS JOIN LATERAL (SELECT ie.item_id FROM item_entities ie
  WHERE ie.entity_id = f.entity_id AND ie.role = f.role AND `+filter+`) ie`,
		append([]any{pgArray(roles), pgArray(ents)}, args...)...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			f  Feature
			it string
		)
		if err := rows.Scan(&f.Role, &f.Entity, &it); err != nil {
			return nil, err
		}

		out[f] = append(out[f], it)
	}

	return out, rows.Err()
}

func currentHashes(ctx context.Context, db *database.DB, ids []string) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT item_id::text, encode(feature_hash, 'hex') FROM item_similar
WHERE item_id = ANY($1::uuid[])`, pgArray(ids))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := map[string]string{}

	for rows.Next() {
		var id, h string
		if err := rows.Scan(&id, &h); err != nil {
			return nil, err
		}

		out[id] = h
	}

	return out, rows.Err()
}

// fillCooccur counts, in one query for the whole chunk, the distinct public lists of at most 100 items that contain
// both a subject and each of its neighbours.
func fillCooccur(ctx context.Context, db *database.DB, subjects []string, kept [][]Scored) error {
	var as, bs []string

	for i, id := range subjects {
		for _, k := range kept[i] {
			as = append(as, id)
			bs = append(bs, k.ID)
		}
	}

	if len(as) == 0 {
		return nil
	}

	rows, err := db.QueryContext(ctx, `SELECT p.a::text, p.b::text, count(DISTINCT x.list_id)
FROM unnest($1::uuid[], $2::uuid[]) AS p(a, b)
JOIN list_items x ON x.item_id = p.a
JOIN lists l ON l.id = x.list_id AND l.is_public AND NOT l.hidden AND l.item_count <= 100
JOIN list_items y ON y.list_id = x.list_id AND y.item_id = p.b
GROUP BY 1, 2`, pgArray(as), pgArray(bs))
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	counts := map[[2]string]int{}

	for rows.Next() {
		var (
			a, b string
			n    int
		)
		if err := rows.Scan(&a, &b, &n); err != nil {
			return err
		}

		counts[[2]string{a, b}] = n
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for i, id := range subjects {
		for j := range kept[i] {
			kept[i][j].Cooccur = counts[[2]string{id, kept[i][j].ID}]
		}
	}

	return nil
}

// featureHash digests F_i, so only a change in the item's own features triggers a rewrite.
func featureHash(f map[Feature]float64) []byte {
	lines := make([]string, 0, len(f))
	for ft, m := range f {
		lines = append(lines, fmt.Sprintf("%s|%s|%g", ft.Role, ft.Entity, m))
	}

	sort.Strings(lines)

	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))

	return sum[:]
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	sort.Strings(out)

	return out
}

// pgArray renders a PostgreSQL array literal bound as a text parameter; every element is quoted and escaped.
func pgArray(vals []string) string {
	esc := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	q := make([]string, len(vals))

	for i, v := range vals {
		q[i] = `"` + esc.Replace(v) + `"`
	}

	return "{" + strings.Join(q, ",") + "}"
}
