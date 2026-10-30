package webhooks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
	vwebhook "github.com/AlexGalhardo/vukem/webhook"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const (
	readLimit  = 10_000
	batchLimit = 100
)

// Fanout reads at most 10,000 api_changes after webhook_cursor, coalesces them
// per item, and enqueues item.batch deliveries. The cursor moves in the same
// transaction as the enqueue, including when no endpoint matches.
func (s *Service) Fanout(ctx context.Context, q *queue.Queue) error {
	return s.db.Tx(ctx, func(tx *database.Tx) error {
		var last int64
		if err := tx.QueryRowContext(ctx, `SELECT last_change_id FROM webhook_cursor WHERE id = 1 FOR UPDATE`).Scan(&last); err != nil {
			return fmt.Errorf("webhook: lock cursor: %w", err)
		}

		rows, err := tx.QueryContext(ctx, `SELECT id, item_id::text, item_type, action, fields, merged_into::text, changed_at
FROM api_changes WHERE id > $1 ORDER BY id LIMIT $2`, last, readLimit)
		if err != nil {
			return fmt.Errorf("webhook: read changes: %w", err)
		}
		defer rows.Close()

		var changes []changeRow
		var maxID int64
		for rows.Next() {
			var row changeRow
			var fields models.TextArray
			var merged sql.NullString
			if err := rows.Scan(&row.id, &row.itemID, &row.itemType, &row.action, &fields, &merged, &row.at); err != nil {
				return err
			}

			row.fields = []string(fields)
			if merged.Valid {
				row.merged = merged.String
			}

			changes = append(changes, row)
			maxID = row.id
		}

		if err := rows.Err(); err != nil {
			return err
		}

		if len(changes) == 0 {
			return nil
		}

		ends, err := loadEndpoints(ctx, tx)
		if err != nil {
			return err
		}

		now := s.now().UTC().Format(time.RFC3339)
		for _, ep := range ends {
			matched := matchEndpoint(ep, coalesce(changes))
			for _, chunk := range chunks(matched, batchLimit) {
				body, err := json.Marshal(batchBody{Type: "item.batch", DeliveredAt: now, Events: chunk})
				if err != nil {
					return err
				}

				id, err := newDeliveryID()
				if err != nil {
					return err
				}

				err = q.DispatchKindTx(ctx, tx, vwebhook.DeliverKind, vwebhook.Delivery{
					ID: id, Endpoint: ep.id, Payload: body,
				})
				if err != nil {
					return fmt.Errorf("webhook: enqueue %s: %w", ep.id, err)
				}
			}
		}

		if _, err := tx.ExecContext(ctx, `UPDATE webhook_cursor SET last_change_id = $1 WHERE id = 1`, maxID); err != nil {
			return fmt.Errorf("webhook: advance cursor: %w", err)
		}

		return nil
	})
}

type changeRow struct {
	id       int64
	itemID   string
	itemType string
	action   string
	fields   []string
	merged   string
	at       time.Time
}

type endpointRow struct {
	id     string
	events []string
	types  []string
}

type batchBody struct {
	Type        string       `json:"type"`
	DeliveredAt string       `json:"deliveredAt"`
	Events      []batchEvent `json:"events"`
}

type batchEvent struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	ItemType   string   `json:"itemType"`
	ItemID     string   `json:"itemId"`
	Fields     []string `json:"fields,omitempty"`
	MergedInto string   `json:"mergedInto,omitempty"`
	ChangedAt  string   `json:"changedAt"`
	seq        int64
}

func loadEndpoints(ctx context.Context, tx *database.Tx) ([]endpointRow, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id::text, events, item_types FROM webhook_endpoints WHERE disabled_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []endpointRow
	for rows.Next() {
		var ep endpointRow
		var events, types models.TextArray
		if err := rows.Scan(&ep.id, &events, &types); err != nil {
			return nil, err
		}

		ep.events = []string(events)
		ep.types = []string(types)
		out = append(out, ep)
	}

	return out, rows.Err()
}

var pluralType = map[string]string{
	"game": "games", "movie": "movies", "series": "series", "book": "books",
}

// coalesce keeps one event per item. A later delete replaces the event.
// A later update after a delete starts a new event. Fields accumulate until
// a delete clears them. The event id is the latest change.
func coalesce(rows []changeRow) []batchEvent {
	byItem := map[string]*batchEvent{}
	var order []string
	for _, row := range rows {
		itemType, ok := pluralType[row.itemType]
		if !ok {
			continue
		}

		ev := byItem[row.itemID]
		if ev == nil {
			ev = &batchEvent{ItemID: row.itemID, ItemType: itemType}
			byItem[row.itemID] = ev
			order = append(order, row.itemID)
		}

		ev.seq = row.id
		ev.ID = fmt.Sprintf("chg_%d", row.id)
		ev.ChangedAt = row.at.UTC().Format(time.RFC3339)
		ev.ItemType = itemType
		switch row.action {
		case "deleted":
			ev.Type = "item.deleted"
			ev.Fields = nil
			ev.MergedInto = row.merged
		case "created":
			ev.Type = "item.created"
			ev.MergedInto = ""
			ev.Fields = unionFields(ev.Fields, row.fields)
		default:
			if ev.Type == "" || ev.Type == "item.deleted" {
				ev.Type = "item.updated"
				ev.Fields = nil
				ev.MergedInto = ""
			}

			ev.Fields = unionFields(ev.Fields, row.fields)
		}
	}

	out := make([]batchEvent, 0, len(order))
	for _, id := range order {
		ev := *byItem[id]
		if ev.Type == "" {
			continue
		}

		slices.Sort(ev.Fields)
		out = append(out, ev)
	}

	slices.SortFunc(out, func(a, b batchEvent) int {
		switch {
		case a.seq < b.seq:
			return -1
		case a.seq > b.seq:
			return 1
		default:
			return 0
		}
	})

	return out
}

func unionFields(got, add []string) []string {
	seen := map[string]struct{}{}
	for _, f := range got {
		seen[f] = struct{}{}
	}

	for _, f := range add {
		if _, ok := seen[f]; ok {
			continue
		}

		seen[f] = struct{}{}
		got = append(got, f)
	}

	return got
}

func matchEndpoint(ep endpointRow, events []batchEvent) []batchEvent {
	var out []batchEvent
	for _, ev := range events {
		if slices.Contains(ep.events, ev.Type) && slices.Contains(ep.types, ev.ItemType) {
			out = append(out, ev)
		}
	}

	return out
}

func chunks(events []batchEvent, n int) [][]batchEvent {
	if len(events) == 0 {
		return nil
	}

	var out [][]batchEvent
	for i := 0; i < len(events); i += n {
		out = append(out, events[i:min(i+n, len(events))])
	}

	return out
}
