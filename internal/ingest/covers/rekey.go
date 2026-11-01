package covers

import (
	"context"
	"fmt"
	"regexp"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// RekeyBatch is the default item page of Rekey.
const RekeyBatch = 500

// RekeyResult counts one Rekey pass: Items looked at, Moved (rows pointed at their slug keys, or that would be
// with a dry run), Copied objects, Current rows already on their slug keys, and Missing rows whose old object is
// not in the bucket (left unchanged).
type RekeyResult struct {
	Items, Moved, Copied, Current, Missing int
}

// hashedKey matches the content-hashed keys Rekey moves (Key, and the legacy Blob pathnames the import kept). Any
// other key, such as a community upload under community/covers/ or a demo SVG, stays where it is.
var hashedKey = regexp.MustCompile(`^covers/[0-9a-f-]+-[0-9a-f]{16}(-t)?\.webp$`)

type rekeyRow struct {
	id, typ, slug, key, thumb string
	thumbTaken                bool
}

// Rekey moves every item cover still under a content-hashed key to its slug key (catalog.CoverKey, owner decision
// 2026-11-01, item 18): it copies the object, then points the row at the copy. It is idempotent: a row already on
// its slug keys is counted as Current, and a re-run after a crash copies again and finishes. The old objects stay;
// the cover_gc trigger files them when the row moves, and covers.gc deletes them 30 days later. A thumbnail whose
// slug key is another item's full key keeps its key (catalog.ThumbCollides). Editions keep theirs: they have no
// slug of their own. Only content-hashed keys move: community uploads and other keys are counted as Current. dryRun copies and updates nothing.
func Rekey(ctx context.Context, db *database.DB, disk storage.Disk, dryRun bool, batch int) (RekeyResult, error) {
	var res RekeyResult

	if disk == nil {
		return res, ErrNoDisk
	}

	if batch <= 0 {
		batch = RekeyBatch
	}

	after := "00000000-0000-0000-0000-000000000000"

	for {
		rows, err := rekeyPage(ctx, db, after, batch)
		if err != nil {
			return res, err
		}

		for _, r := range rows {
			res.Items++

			if err := rekeyOne(ctx, db, disk, r, dryRun, &res); err != nil {
				return res, fmt.Errorf("covers rekey %s: %w", r.id, err)
			}
		}

		if len(rows) < batch {
			return res, nil
		}

		after = rows[len(rows)-1].id
	}
}

func rekeyPage(ctx context.Context, db *database.DB, after string, limit int) ([]rekeyRow, error) {
	rows, err := db.QueryContext(ctx, `SELECT i.id::text, i.type::text, i.slug, coalesce(i.cover_key, ''),
  coalesce(i.cover_thumb_key, ''), EXISTS (SELECT 1 FROM items o WHERE o.type = i.type AND o.slug = i.slug || '-t')
FROM items i
WHERE i.id > $1::uuid AND (i.cover_key IS NOT NULL OR i.cover_thumb_key IS NOT NULL)
ORDER BY i.id LIMIT $2`, after, limit)
	if err != nil {
		return nil, fmt.Errorf("covers rekey: list items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []rekeyRow

	for rows.Next() {
		var r rekeyRow
		if err := rows.Scan(&r.id, &r.typ, &r.slug, &r.key, &r.thumb, &r.thumbTaken); err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, rows.Err()
}

func rekeyOne(
	ctx context.Context,
	db *database.DB,
	disk storage.Disk,
	r rekeyRow,
	dryRun bool,
	res *RekeyResult,
) error {
	key, thumb := r.key, r.thumb
	if hashedKey.MatchString(key) {
		key = catalog.CoverKey(r.typ, r.slug, false)
	}

	if hashedKey.MatchString(thumb) && !r.thumbTaken {
		thumb = catalog.CoverKey(r.typ, r.slug, true)
	}

	if key == r.key && thumb == r.thumb {
		res.Current++
		return nil
	}

	// Both objects must exist before the row moves, or the page would point at a missing key.
	for _, p := range [][2]string{{r.key, key}, {r.thumb, thumb}} {
		if p[0] == p[1] {
			continue
		}

		ok, err := disk.Exists(ctx, p[0])
		if err != nil {
			return err
		}

		if !ok {
			res.Missing++
			return nil
		}
	}

	res.Moved++

	if dryRun {
		return nil
	}

	for _, p := range [][2]string{{r.key, key}, {r.thumb, thumb}} {
		if p[0] == p[1] {
			continue
		}

		if err := copyObject(ctx, disk, p[0], p[1]); err != nil {
			return err
		}

		res.Copied++
	}

	// Guarded by the old keys: a covers run that replaced the cover meanwhile wins, and its keys are already slug keys.
	_, err := db.ExecContext(ctx, `UPDATE items SET cover_key = nullif($2, ''), cover_thumb_key = nullif($3, '')
WHERE id = $1 AND coalesce(cover_key, '') = $4 AND coalesce(cover_thumb_key, '') = $5`,
		r.id, key, thumb, r.key, r.thumb)

	return err
}

func copyObject(ctx context.Context, disk storage.Disk, from, to string) error {
	f, err := disk.Open(ctx, from)
	if err != nil {
		return fmt.Errorf("copy %s: %w", from, err)
	}
	defer func() { _ = f.Close() }()

	if err := disk.Put(ctx, to, f); err != nil {
		return fmt.Errorf("copy %s to %s: %w", from, to, err)
	}

	return nil
}
