//go:build integration

package legacyimport

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// This file is the per-item implementation the entities stage had before the set-based rewrite
// (FIX-legacyimport-entities-setbased), kept verbatim as the oracle the equivalence tests compare with.

// referenceEntitiesBatch turns the next n legacy items' meta strings into entities and item_entities, and their
// isbn13 into one edition (SDD 01 §5.4, open issue 01-1). It is the minimal mapping; the Go normaliser
// (P6.A) rebuilds entities from source_records later.
func referenceEntitiesBatch() BatchFunc {
	// The cache spans batches; it is dropped whenever a batch fails, since its rows roll back.
	ids := map[string]string{}

	return func(ctx context.Context, tx *database.Tx, after string, n int) (last string, rows int, err error) {
		defer func() {
			if err != nil {
				clear(ids)
			}
		}()

		if after == "" {
			if _, err := tx.ExecContext(ctx, legacySource); err != nil {
				return "", 0, fmt.Errorf("entities seed: %w", err)
			}
		}

		items, err := loadEntityItems(ctx, tx, after, n)
		if err != nil || len(items) == 0 {
			return "", 0, err
		}

		r := refResolver{tx: tx, ids: ids}

		for _, it := range items {
			if err := r.link(ctx, it); err != nil {
				return "", 0, fmt.Errorf("entities item %s: %w", it.id, err)
			}

			if err := refAddEdition(ctx, tx, it); err != nil {
				return "", 0, fmt.Errorf("entities edition %s: %w", it.id, err)
			}
		}

		return items[len(items)-1].id, len(items), nil
	}
}

// refResolver finds or creates entities; ids caches the lookups (kind, QID, name) of committed batches.
type refResolver struct {
	tx  *database.Tx
	ids map[string]string
}

func (r *refResolver) link(ctx context.Context, it entityItem) error {
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

			id, err := r.entity(ctx, l.mr.kind, qid, name)
			if err != nil {
				return err
			}

			if _, err := r.tx.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`, it.id, id, l.mr.role, pos, source); err != nil {
				return err
			}
		}
	}

	return nil
}

// entity returns the entity with that QID (whatever its kind), else the native entity of that kind and
// name, else a new one; native entities (no QID) are what DQ-09 lists for review.
func (r *refResolver) entity(ctx context.Context, kind, qid, name string) (string, error) {
	key := kind + "\x00" + qid + "\x00" + name
	if id, ok := r.ids[key]; ok {
		return id, nil
	}

	var (
		id  string
		err error
	)
	if qid != "" {
		err = r.tx.QueryRowContext(ctx, `SELECT id::text FROM entities WHERE wikidata_qid = $1`, qid).Scan(&id)
	} else {
		err = r.tx.QueryRowContext(ctx, `SELECT id::text FROM entities
WHERE wikidata_qid IS NULL AND kind = $1 AND default_name = $2 LIMIT 1`, kind, name).Scan(&id)
	}
	if errors.Is(err, sql.ErrNoRows) {
		id, err = r.create(ctx, kind, qid, name)
	}

	if err != nil {
		return "", err
	}

	r.ids[key] = id

	return id, nil
}

func (r *refResolver) create(ctx context.Context, kind, qid, name string) (string, error) {
	slug, err := catalog.UniqueSlug(ctx, r.tx, catalog.ScopeEntity, kind, name, 0)
	if errors.Is(err, catalog.ErrEmptySlug) {
		// A name without Latin letters or digits (e.g. 日本): slug from the QID, else a name hash.
		base := strings.ToLower(qid)
		if base == "" {
			sum := sha256.Sum256([]byte(name))
			base = kind + "-" + hex.EncodeToString(sum[:4])
		}

		slug, err = catalog.UniqueSlug(ctx, r.tx, catalog.ScopeEntity, kind, base, 0)
	}

	if err != nil {
		return "", err
	}

	id := uuid.NewV7().String()

	_, err = r.tx.ExecContext(ctx, `INSERT INTO entities (id, kind, slug, wikidata_qid, default_name)
VALUES ($1, $2, $3, nullif($4, ''), $5)`, id, kind, slug, qid, name)

	return id, err
}

// refAddEdition creates one edition per legacy book isbn13 with an unknown language (DQ-10). A book whose
// ISBN is missing or invalid, or already held by another work, simply gets no edition (01-1).
func refAddEdition(ctx context.Context, tx *database.Tx, it entityItem) error {
	isbn := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}

		return -1
	}, it.isbn)
	if it.typ != "book" || !catalog.ValidISBN13(isbn) {
		return nil
	}

	_, err := tx.ExecContext(ctx, `INSERT INTO editions (id, work_id, isbn13, title, language, source)
VALUES ($1, $2, $3, $4, 'und', 'legacy') ON CONFLICT (isbn13) DO NOTHING`, uuid.NewV7().String(), it.id, isbn, it.title)

	return err
}
