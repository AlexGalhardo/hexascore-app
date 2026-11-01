package normalize

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
)

// publicSnapshot reads the API-visible values ingestion writes, one JSON text per API field name.
func publicSnapshot(ctx context.Context, ex database.Executor, id string) (map[string]string, error) {
	var raw []byte

	err := ex.QueryRowContext(ctx, `SELECT jsonb_build_object(
  'slug', slug, 'wikidata', wikidata_qid, 'releaseDate', release_date, 'runtimeMinutes', runtime_minutes,
  'pageCount', page_count, 'externalIds', external_ids,
  'title', (SELECT coalesce(jsonb_object_agg(locale, title), '{}') FROM item_translations WHERE item_id = items.id),
  'synopsis', (SELECT coalesce(jsonb_object_agg(locale, synopsis), '{}') FROM item_translations WHERE item_id = items.id))
FROM items WHERE id = $1`, id).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("snapshot item %s: %w", id, err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}

	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = string(v)
	}

	return out, nil
}

// recordChange writes the api_changes row of a saved item: created when before is nil, otherwise updated with the
// public fields whose value differs from before (nothing is written when none does).
func recordChange(ctx context.Context, ex database.Executor, id, kind string, before map[string]string) error {
	if before == nil {
		return changes.Record(ctx, ex, changes.Change{ItemID: id, ItemType: kind, Action: "created"})
	}

	after, err := publicSnapshot(ctx, ex, id)
	if err != nil {
		return err
	}

	var fields []string

	for k, v := range after {
		if before[k] != v {
			fields = append(fields, k)
		}
	}

	slices.Sort(fields)

	return changes.Record(ctx, ex, changes.Change{ItemID: id, ItemType: kind, Action: "updated", Fields: fields})
}

// withEntitiesChange runs save and, when it changed the item's credits, writes an `updated` row naming `entities`.
func withEntitiesChange(ctx context.Context, ex database.Executor, id, kind string, save func() error) error {
	before, err := entitiesFingerprint(ctx, ex, id)
	if err != nil {
		return err
	}

	if err := save(); err != nil {
		return err
	}

	after, err := entitiesFingerprint(ctx, ex, id)
	if err != nil || after == before {
		return err
	}

	return changes.Record(
		ctx,
		ex,
		changes.Change{ItemID: id, ItemType: kind, Action: "updated", Fields: []string{"entities"}},
	)
}

func entitiesFingerprint(ctx context.Context, ex database.Executor, id string) (string, error) {
	var fp string

	err := ex.QueryRowContext(ctx, `SELECT coalesce(string_agg(role || ':' || entity_id::text || ':' || position
  || ':' || coalesce(character_name, ''), ',' ORDER BY role, entity_id), '') FROM item_entities WHERE item_id = $1`, id).
		Scan(&fp)
	if err != nil {
		return "", fmt.Errorf("fingerprint credits of %s: %w", id, err)
	}

	return fp, nil
}
