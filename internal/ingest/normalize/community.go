package normalize

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// SourceCommunity is the data_sources code of approved community edits.
const SourceCommunity = "community"

// CommunityLock is how long a connector skips a community field after the
// applying audit row (FR-06-4, SDD 04 §7.3, open issue 04-3).
const CommunityLock = 180 * 24 * time.Hour

// CommunityField is a field whose latest applied change is a community edit.
type CommunityField struct {
	// Since is the occurred_at of the applying audit row.
	Since time.Time
	// Current is the live value, encoded like Item.FieldValue.
	Current json.RawMessage
}

// SystemSuggestion asks the moderators to look at a source value that differs
// from an approved community value. Only an admin can apply it (SDD 04).
type SystemSuggestion struct {
	ItemID, Domain, Field string
	Before, After         json.RawMessage
	// Source is the data_sources code that supplied After.
	Source string
}

// Community is the SDD 04 side of the merge, implemented by app/community.
// The normaliser never writes suggestion or audit tables itself.
type Community interface {
	CommunityFields(ctx context.Context, itemID string) (map[string]CommunityField, error)
	OpenSystemSuggestion(ctx context.Context, s SystemSuggestion) error
}

// CommunityResult lists what ApplyCommunity did.
type CommunityResult struct {
	Skipped, Suggested []string
}

// FieldValue encodes the merged value of one SDD 04 §7.2 field name; ok is
// false when the item carries no value for it.
func (it *Item) FieldValue(field string) (json.RawMessage, bool) {
	var v any

	switch {
	case field == "release_date" && it.ReleaseDate != nil:
		v = map[string]string{"value": it.ReleaseDate.Format("2006-01-02"), "precision": it.ReleasePrecision}
	case strings.HasPrefix(field, "external_ids."):
		v = it.ExternalIDs[strings.TrimPrefix(field, "external_ids.")]
	case strings.HasPrefix(field, "synopsis."):
		v = it.Synopses[strings.TrimPrefix(field, "synopsis.")].Text
	case strings.HasPrefix(field, "translations.") && strings.HasSuffix(field, ".title"):
		v = it.Titles[strings.TrimSuffix(strings.TrimPrefix(field, "translations."), ".title")]
	default:
		return nil, false
	}

	if s, isString := v.(string); isString && s == "" {
		return nil, false
	}

	raw, err := json.Marshal(v)

	return raw, err == nil
}

// ApplyCommunity keeps ingestion away from community-owned fields. A field
// whose value the source also has is skipped for CommunityLock after the
// edit; a value that differs opens a system suggestion and is never written,
// before or after the lock expires. Skipped fields are marked in item.Skip and
// dropped from item.Provenance; the writer must not store them.
func ApplyCommunity(
	ctx context.Context,
	c Community,
	itemID string,
	item *Item,
	now time.Time,
) (CommunityResult, error) {
	var res CommunityResult

	fields, err := c.CommunityFields(ctx, itemID)
	if err != nil {
		return res, fmt.Errorf("read community fields of %s: %w", itemID, err)
	}

	for field, cf := range fields {
		after, ok := item.FieldValue(field)
		if !ok {
			continue
		}

		diverges := !jsonEqual(after, cf.Current)
		locked := now.Sub(cf.Since) < CommunityLock

		if diverges {
			s := SystemSuggestion{
				ItemID: itemID, Domain: item.Type, Field: field,
				Before: cf.Current, After: after, Source: item.Provenance[field],
			}
			if err := c.OpenSystemSuggestion(ctx, s); err != nil {
				return res, fmt.Errorf("open system suggestion for %s %s: %w", itemID, field, err)
			}

			res.Suggested = append(res.Suggested, field)
		}

		if locked || diverges {
			item.Skip[field] = true
			delete(item.Provenance, field)

			res.Skipped = append(res.Skipped, field)
		}
	}

	return res, nil
}

func jsonEqual(a, b json.RawMessage) bool {
	var x, y any
	if json.Unmarshal(a, &x) != nil || json.Unmarshal(b, &y) != nil {
		return false
	}

	return reflect.DeepEqual(x, y)
}
