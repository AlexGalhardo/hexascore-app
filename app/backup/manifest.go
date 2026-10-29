package backup

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// Manifest is the row count of every table in schema public, taken inside the
// snapshot pg_dump exports, so a restore of the dump must reproduce it exactly
// (SDD 11 §7.4). It is stored next to the dump, sealed with the same key.
type Manifest struct {
	Key    string           `json:"key"`
	Tables map[string]int64 `json:"tables"`
}

// countRowsSQL counts every public table in one statement; query_to_xml runs
// each count inside the caller's transaction, so all share its snapshot.
const countRowsSQL = `SELECT tablename::text,
  (xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM public.%I', tablename), false, true, '')))[1]::text::bigint
FROM pg_tables WHERE schemaname = 'public'`

type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// CountRows returns the row count of every table in schema public.
func CountRows(ctx context.Context, q querier) (map[string]int64, error) {
	rows, err := q.QueryContext(ctx, countRowsSQL)
	if err != nil {
		return nil, fmt.Errorf("backup: count rows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	counts := map[string]int64{}

	for rows.Next() {
		var (
			table string
			n     int64
		)

		if err := rows.Scan(&table, &n); err != nil {
			return nil, fmt.Errorf("backup: count rows: %w", err)
		}

		counts[table] = n
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("backup: count rows: %w", err)
	}

	return counts, nil
}

// CountDiff lists the tables whose restored count differs from the manifest,
// including tables missing on either side.
func CountDiff(want, got map[string]int64) []string {
	var diff []string

	for _, t := range slices.Sorted(maps.Keys(want)) {
		if n, ok := got[t]; !ok || n != want[t] {
			diff = append(diff, fmt.Sprintf("%s: want %d, got %d", t, want[t], n))
		}
	}

	for _, t := range slices.Sorted(maps.Keys(got)) {
		if _, ok := want[t]; !ok {
			diff = append(diff, fmt.Sprintf("%s: not in the manifest, got %d", t, got[t]))
		}
	}

	return diff
}

// ManifestKey is the manifest object of a dump. It lives under its own prefix
// so listing and pruning backups/postgres/<kind> sees dumps only.
func ManifestKey(dumpKey string) string {
	return strings.Replace(dumpKey, "backups/postgres/", "backups/manifests/postgres/", 1) + ".json"
}

func (r *Runner) putManifest(ctx context.Context, m Manifest) error {
	raw, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("backup: manifest: %w", err)
	}

	var sealed bytes.Buffer

	key := ManifestKey(m.Key)
	fw := NewFrameWriter(&sealed, r.Enc, key)

	if _, err := fw.Write(raw); err != nil {
		return fmt.Errorf("backup: seal manifest: %w", err)
	}

	if err := fw.Close(); err != nil {
		return fmt.Errorf("backup: seal manifest: %w", err)
	}

	if err := r.Disk.Put(ctx, key, &sealed); err != nil {
		return fmt.Errorf("backup: upload manifest: %w", err)
	}

	return nil
}

// ReadManifest opens the manifest of the dump at dumpKey.
func (r *Runner) ReadManifest(ctx context.Context, dumpKey string) (Manifest, error) {
	key := ManifestKey(dumpKey)

	obj, err := r.Disk.Open(ctx, key)
	if err != nil {
		return Manifest{}, fmt.Errorf("backup: open manifest: %w", err)
	}
	defer func() { _ = obj.Close() }()

	raw, err := io.ReadAll(NewFrameReader(obj, r.Enc, key))
	if err != nil {
		return Manifest{}, fmt.Errorf("backup: read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, fmt.Errorf("backup: decode manifest: %w", err)
	}

	if m.Key != dumpKey {
		return Manifest{}, fmt.Errorf("backup: manifest names %q, want %q", m.Key, dumpKey)
	}

	return m, nil
}
