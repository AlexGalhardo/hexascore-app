package legacyimport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"slices"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// legacySource marks rows whose origin the legacy dump does not record (meta strings that did not come
// from Wikidata, legacy ISBNs). Its licence is unknown, so the API never serves them (CATALOG_API_SOURCES).
const legacySource = `INSERT INTO data_sources (code, name, licence, attribution, url) VALUES
  ('legacy', 'HexaScore legacy import', 'unknown', NULL, 'https://hexascore.app')
ON CONFLICT (code) DO NOTHING`

// metaRole maps a legacy meta list to its role, entity kind and Wikidata property (SDD 01 §5.2 role
// table, SDD 06 §7.4, legacy normalize.py). Legacy creators carry no role: they get one per item type.
type metaRole struct{ role, kind, prop string }

var creatorRoles = map[string]metaRole{
	"movie":  {"creator", "person", "P57"},
	"series": {"creator", "person", "P170"},
	"game":   {"developer", "organization", "P178"},
	"book":   {"author", "person", "P50"},
}

type legacyMeta struct {
	Genres    []string `json:"genres"`
	Platforms []string `json:"platforms"`
	Creators  []string `json:"creators"`
	Countries []string `json:"countries"`
}

type wikidataClaims struct {
	Claims map[string][]struct {
		Value any `json:"value"`
	} `json:"claims"`
	Refs map[string]string `json:"refs"`
}

// qid returns the Wikidata id behind a meta string: a claim of prop whose referenced label is name
// (legacy normalize.py ref_labels built the strings exactly this way).
func (w wikidataClaims) qid(prop, name string) string {
	for _, c := range w.Claims[prop] {
		if id, ok := c.Value.(string); ok && w.Refs[id] == name {
			return id
		}
	}

	return ""
}

type entityItem struct {
	id, typ, isbn, title string
	meta                 legacyMeta
	wd                   wikidataClaims
}

// entityKeysSQL lists the item id ending every batch of $2 items after $1 (Stage.KeysSQL).
const entityKeysSQL = `SELECT id::text FROM (SELECT id, row_number() OVER (ORDER BY id) AS rn FROM legacy_import.items
WHERE id > ` + afterKey + `) x WHERE rn % $2 = 0 ORDER BY id`

// entityRef names the entity a meta string resolves to: the one holding qid (whatever its kind), else
// the native entity of that kind and name.
type entityRef struct{ kind, qid, name string }

const keySep = "\x00"

func (e entityRef) key() string {
	if e.qid != "" {
		return "q" + keySep + e.qid
	}

	return "n" + keySep + e.kind + keySep + e.name
}

type entityLink struct {
	itemID, role, source string
	ref                  entityRef
	pos                  int
}

// entitiesBatch turns the next n legacy items' meta strings into entities and item_entities, and their
// isbn13 into one edition (SDD 01 §5.4, open issue 01-1). It is the minimal mapping; the Go normaliser
// (P6.A) rebuilds entities from source_records later. A batch costs a fixed handful of statements
// whatever its size: one read of the items, a lookup of the entities, one insert of the missing ones and
// one insert each of the links and the editions. It is safe to run batches of disjoint item ranges
// concurrently (Runner.Workers).
func entitiesBatch() BatchFunc {
	return func(ctx context.Context, tx *database.Tx, after string, n int) (string, int, error) {
		items, err := loadEntityItems(ctx, tx, after, n)
		if err != nil || len(items) == 0 {
			return "", 0, err
		}

		// Every batch seeds it: parallel workers cannot count on the first batch having committed.
		if _, err := tx.ExecContext(ctx, legacySource); err != nil {
			return "", 0, fmt.Errorf("entities seed: %w", err)
		}

		links := collectLinks(items)

		ids, err := resolveEntities(ctx, tx, links)
		if err != nil {
			return "", 0, fmt.Errorf("entities resolve: %w", err)
		}

		if err := insertLinks(ctx, tx, links, ids); err != nil {
			return "", 0, fmt.Errorf("entities links: %w", err)
		}

		if err := insertEditions(ctx, tx, items); err != nil {
			return "", 0, fmt.Errorf("entities editions: %w", err)
		}

		return items[len(items)-1].id, len(items), nil
	}
}

func loadEntityItems(ctx context.Context, tx *database.Tx, after string, n int) ([]entityItem, error) {
	rows, err := tx.QueryContext(ctx, `SELECT i.id::text, i.type::text, i.meta, coalesce(i.external_ids->>'isbn13', ''),
  coalesce(s.payload, '{}'),
  coalesce((SELECT t.title FROM item_translations t WHERE t.item_id = i.id ORDER BY t.locale <> 'en', t.locale LIMIT 1), i.slug)
FROM legacy_import.items i
LEFT JOIN legacy_import.source_records s ON s.item_id = i.id AND s.source = 'wikidata'
WHERE i.id > `+afterKey+` ORDER BY i.id LIMIT $2`, after, n)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []entityItem

	for rows.Next() {
		var (
			it          entityItem
			meta, claim []byte
		)

		if err := rows.Scan(&it.id, &it.typ, &meta, &it.isbn, &claim, &it.title); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(meta, &it.meta); err != nil {
			return nil, fmt.Errorf("item %s meta: %w", it.id, err)
		}

		if err := json.Unmarshal(claim, &it.wd); err != nil {
			return nil, fmt.Errorf("item %s wikidata payload: %w", it.id, err)
		}

		out = append(out, it)
	}

	return out, rows.Err()
}

func collectLinks(items []entityItem) []entityLink {
	var out []entityLink

	for _, it := range items {
		lists := []struct {
			names []string
			mr    metaRole
		}{
			{it.meta.Genres, metaRole{"genre", "genre", "P136"}},
			{it.meta.Platforms, metaRole{"platform", "platform", "P400"}},
			{it.meta.Creators, creatorRoles[it.typ]},
			{it.meta.Countries, metaRole{"country", "country", "P495"}},
		}

		for _, l := range lists {
			for pos, name := range l.names {
				name = strings.TrimSpace(name)
				if name == "" || l.mr.role == "" {
					continue
				}

				qid, source := it.wd.qid(l.mr.prop, name), "wikidata"
				if qid == "" {
					source = "legacy"
				}

				out = append(out, entityLink{
					itemID: it.id, role: l.mr.role, source: source, pos: pos,
					ref: entityRef{kind: l.mr.kind, qid: qid, name: name},
				})
			}
		}
	}

	return out
}

// resolveEntities returns the id of every entity the links name, keyed by entityRef.key, creating the
// missing ones. Two workers may race to create the same entity: the creator takes advisory locks on the
// QID and on the kind+slug first, so the other waits for its commit, finds the row and creates nothing.
func resolveEntities(ctx context.Context, tx *database.Tx, links []entityLink) (map[string]string, error) {
	ids := map[string]string{}

	var refs []entityRef

	for _, l := range links {
		if _, dup := ids[l.ref.key()]; !dup {
			ids[l.ref.key()] = ""

			refs = append(refs, l.ref)
		}
	}

	missing, err := lookupEntities(ctx, tx, refs, ids)
	if err != nil || len(missing) == 0 {
		return ids, err
	}

	if err := lockEntities(ctx, tx, missing); err != nil {
		return nil, err
	}

	if missing, err = lookupEntities(ctx, tx, missing, ids); err != nil || len(missing) == 0 {
		return ids, err
	}

	slugs, err := assignSlugs(ctx, tx, missing)
	if err != nil {
		return nil, err
	}

	var newIDs, kinds, qids, names []string

	for _, m := range missing {
		newIDs, kinds = append(newIDs, uuid.NewV7().String()), append(kinds, m.kind)
		qids, names = append(qids, m.qid), append(names, m.name)
	}

	res, err := tx.ExecContext(ctx, `INSERT INTO entities (id, kind, slug, wikidata_qid, default_name)
SELECT t.id::uuid, t.kind, t.slug, nullif(t.qid, ''), t.name
FROM unnest($1::text[], $2::text[], $3::text[], $4::text[], $5::text[]) AS t(id, kind, slug, qid, name)
ON CONFLICT DO NOTHING`, newIDs, kinds, slugs, qids, names)
	if err != nil {
		return nil, err
	}

	if inserted, _ := res.RowsAffected(); int(inserted) == len(missing) {
		for i, m := range missing {
			ids[m.key()] = newIDs[i]
		}

		return ids, nil
	}

	// A concurrent creator won some of them between the lookup and the insert.
	if missing, err = lookupEntities(ctx, tx, missing, ids); err != nil {
		return nil, err
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("entity %s %q not created (slug taken concurrently); rerun the stage",
			missing[0].kind, missing[0].name)
	}

	return ids, nil
}

// lookupEntities fills ids for the refs that exist and returns those that do not: the entity with that
// QID (whatever its kind), else the native entity of that kind and name; native entities (no QID) are
// what DQ-09 lists for review.
func lookupEntities(
	ctx context.Context,
	tx *database.Tx,
	refs []entityRef,
	ids map[string]string,
) ([]entityRef, error) {
	var qids, kinds, names []string

	for _, r := range refs {
		if r.qid != "" {
			qids = append(qids, r.qid)
		} else {
			kinds, names = append(kinds, r.kind), append(names, r.name)
		}
	}

	if len(qids) > 0 {
		err := scanIDs(ctx, tx, func(k, id string) { ids["q"+keySep+k] = id },
			`SELECT wikidata_qid, id::text FROM entities WHERE wikidata_qid = ANY($1::text[])`, qids)
		if err != nil {
			return nil, err
		}
	}

	if len(kinds) > 0 {
		err := scanIDs(ctx, tx, func(k, id string) { ids["n"+keySep+k] = id },
			`SELECT DISTINCT ON (e.kind, e.default_name) e.kind || chr(1) || e.default_name, e.id::text
FROM entities e JOIN unnest($1::text[], $2::text[]) n(kind, name) ON e.kind = n.kind AND e.default_name = n.name
WHERE e.wikidata_qid IS NULL ORDER BY e.kind, e.default_name, e.id`, kinds, names)
		if err != nil {
			return nil, err
		}
	}

	var missing []entityRef

	for _, r := range refs {
		if ids[r.key()] == "" {
			missing = append(missing, r)
		}
	}

	return missing, nil
}

// scanIDs runs a query of (key, id) rows; a native key arrives as kind + chr(1) + name.
func scanIDs(ctx context.Context, tx *database.Tx, set func(key, id string), query string, args ...any) error {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var k, id string
		if err := rows.Scan(&k, &id); err != nil {
			return err
		}

		set(strings.Replace(k, "\x01", keySep, 1), id)
	}

	return rows.Err()
}

// slugBase is the slug UniqueSlug starts from: the name, else (no Latin letters or digits, e.g. 日本)
// the QID, else a name hash.
func slugBase(m entityRef) string {
	if base := catalog.Slugify(m.name); base != "" {
		return base
	}

	base := strings.ToLower(m.qid)
	if base == "" {
		sum := sha256.Sum256([]byte(m.name))
		base = m.kind + "-" + hex.EncodeToString(sum[:4])
	}

	return catalog.Slugify(base)
}

func lockEntities(ctx context.Context, tx *database.Tx, missing []entityRef) error {
	seen := map[int64]bool{}
	locks := make([]int64, 0, 2*len(missing))

	add := func(s string) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(s))

		if k := int64(h.Sum32()); !seen[k] {
			seen[k] = true

			locks = append(locks, k)
		}
	}

	for _, m := range missing {
		add("legacyimport.entities" + keySep + m.kind + keySep + slugBase(m))

		if m.qid != "" {
			add("legacyimport.entities" + keySep + m.qid)
		}
	}

	// Sorted and taken by one statement: every worker takes the same locks in the same order, so none deadlocks.
	slices.Sort(locks)

	_, err := tx.ExecContext(ctx,
		`SELECT count(pg_advisory_xact_lock(u.k)) FROM unnest($1::bigint[]) WITH ORDINALITY AS u(k, o)`, locks)

	return err
}

// assignSlugs gives each missing entity, in order, the slug UniqueSlug would have chosen had they been
// created one by one: the slugs already taken (entities and redirects) are read in one query, the ones
// handed out earlier in the batch are tracked here.
func assignSlugs(ctx context.Context, tx *database.Tx, missing []entityRef) ([]string, error) {
	var kinds, exact, pats []string

	bases := make([]string, len(missing))

	for i, m := range missing {
		bases[i] = slugBase(m)

		// Candidates are base-N, with base cut short only near MaxSlugLength, so a prefix covers them.
		pat := bases[i] + "-%"
		if room := catalog.MaxSlugLength - 8; len(bases[i]) > room {
			pat = bases[i][:room] + "%"
		}

		kinds, exact, pats = append(kinds, m.kind), append(exact, bases[i]), append(pats, pat)
	}

	taken := map[string]bool{}

	rows, err := tx.QueryContext(ctx, `SELECT e.kind, e.slug FROM entities e
JOIN unnest($1::text[], $2::text[], $3::text[]) p(kind, exact, pat)
  ON e.kind = p.kind AND (e.slug = p.exact OR e.slug LIKE p.pat)
UNION SELECT c.namespace, c.old_slug FROM catalog_redirects c
JOIN unnest($1::text[], $2::text[], $3::text[]) p(kind, exact, pat)
  ON c.scope = 'entity' AND c.namespace = p.kind AND (c.old_slug = p.exact OR c.old_slug LIKE p.pat)`, kinds, exact, pats)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var kind, slug string
		if err := rows.Scan(&kind, &slug); err != nil {
			return nil, err
		}

		taken[kind+keySep+slug] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]string, len(missing))

	for i, m := range missing {
		out[i] = catalog.FirstFreeSlug(bases[i], func(s string) bool { return taken[m.kind+keySep+s] })
		taken[m.kind+keySep+out[i]] = true
	}

	return out, nil
}

func insertLinks(ctx context.Context, tx *database.Tx, links []entityLink, ids map[string]string) error {
	if len(links) == 0 {
		return nil
	}

	var itemIDs, entityIDs, roles, sources []string

	positions := make([]int, 0, len(links))

	for _, l := range links {
		itemIDs, entityIDs = append(itemIDs, l.itemID), append(entityIDs, ids[l.ref.key()])
		roles, sources, positions = append(roles, l.role), append(sources, l.source), append(positions, l.pos)
	}

	_, err := tx.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
SELECT t.item_id::uuid, t.entity_id::uuid, t.role, t.position::smallint, t.source
FROM unnest($1::text[], $2::text[], $3::text[], $4::int[], $5::text[]) AS t(item_id, entity_id, role, position, source)
ON CONFLICT DO NOTHING`, itemIDs, entityIDs, roles, positions, sources)

	return err
}

// insertEditions creates one edition per legacy book isbn13 with an unknown language (DQ-10). A book whose
// ISBN is missing or invalid, or already held by another work, simply gets no edition (01-1).
func insertEditions(ctx context.Context, tx *database.Tx, items []entityItem) error {
	var edIDs, works, isbns, titles []string

	for _, it := range items {
		isbn := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}

			return -1
		}, it.isbn)
		if it.typ != "book" || !catalog.ValidISBN13(isbn) {
			continue
		}

		edIDs, works = append(edIDs, uuid.NewV7().String()), append(works, it.id)
		isbns, titles = append(isbns, isbn), append(titles, it.title)
	}

	if len(works) == 0 {
		return nil
	}

	_, err := tx.ExecContext(ctx, `INSERT INTO editions (id, work_id, isbn13, title, language, source)
SELECT t.id::uuid, t.work_id::uuid, t.isbn13, t.title, 'und', 'legacy'
FROM unnest($1::text[], $2::text[], $3::text[], $4::text[]) AS t(id, work_id, isbn13, title)
ON CONFLICT (isbn13) DO NOTHING`, edIDs, works, isbns, titles)

	return err
}
