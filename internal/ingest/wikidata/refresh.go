package wikidata

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"slices"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// EntitySource is the source_records source of referenced entities (people, companies, genres...).
const EntitySource = "wikidata:entity"

// Entities refreshes the "wikidata:entity" records already stored. New entities arrive with the items that
// reference them (Connector.Fetch), so only a full run has work here.
type Entities struct {
	w *Connector
}

var _ ingest.Connector = Entities{}

// Entities returns the entity refresher sharing this connector's endpoints.
func (w *Connector) Entities() Entities { return Entities{w: w} }

// Source names the connector in source_records.
func (Entities) Source() string { return EntitySource }

// Targets yields the stored entity ids after the cursor in full mode, and nothing in any other mode.
func (Entities) Targets(
	ctx context.Context, db database.Executor, p ingest.RunParams, after string,
) iter.Seq2[ingest.Target, error] {
	return func(yield func(ingest.Target, error) bool) {
		if p.Mode != run.ModeFull {
			return
		}

		rows, err := db.QueryContext(ctx,
			`SELECT external_id FROM source_records WHERE source = $1 AND external_id > $2 ORDER BY external_id`,
			EntitySource, after)
		if err != nil {
			yield(ingest.Target{}, fmt.Errorf("wikidata entity targets: %w", err))

			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				yield(ingest.Target{}, fmt.Errorf("scan wikidata entity target: %w", err))

				return
			}

			if !yield(ingest.Target{ID: id}, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(ingest.Target{}, fmt.Errorf("wikidata entity targets: %w", err))
		}
	}
}

// Fetch re-reads the entities in pages of Batch. An id the API no longer returns keeps its stored record.
func (e Entities) Fetch(
	ctx context.Context, c *ingest.Client, batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	var (
		records []ingest.Record
		errs    []ingest.TargetError
	)

	for chunk := range slices.Chunk(batch, Batch) {
		ids := make([]string, len(chunk))
		for i, t := range chunk {
			ids[i] = t.ID
		}

		raws, err := e.w.getEntities(ctx, c, ids)
		if err != nil {
			for _, t := range chunk {
				errs = append(errs, ingest.TargetError{Target: t, Err: err})
			}

			continue
		}

		for _, t := range chunk {
			raw, ok := raws[t.ID]
			if !ok {
				continue
			}

			ent, err := ParseEntity(raw)
			if err != nil {
				errs = append(errs, ingest.TargetError{Target: t, Err: err})

				continue
			}

			if ent.Missing != nil {
				continue
			}

			body, err := json.Marshal(TrimEntity(ent))
			if err != nil {
				errs = append(errs, ingest.TargetError{Target: t, Err: fmt.Errorf("encode entity %s: %w", t.ID, err)})

				continue
			}

			records = append(records, ingest.Record{Source: EntitySource, ExternalID: t.ID, Payload: body})
		}
	}

	return records, errs
}
