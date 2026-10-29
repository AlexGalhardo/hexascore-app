package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
)

// Scope tells whether a redirect row concerns items or entities.
type Scope string

// Redirect scopes, as stored in catalog_redirects.scope.
const (
	ScopeItem   Scope = "item"
	ScopeEntity Scope = "entity"
)

// ErrSlugReserved means the slug is a redirect old_slug of another row (I-1).
var ErrSlugReserved = errors.New("catalog: slug is reserved by a redirect to another row")

// Target is the live row a redirect resolves to.
type Target struct {
	ID, Slug string
}

// table returns the table and namespace column of the scope; the values are
// constants, never user input, so they may be formatted into SQL.
func (s Scope) table() (table, col string, err error) {
	switch s {
	case ScopeItem:
		return "items", "type", nil
	case ScopeEntity:
		return "entities", "kind", nil
	default:
		return "", "", fmt.Errorf("catalog: unknown redirect scope %q", s)
	}
}

// RenameSlug changes the slug of a live row and records the old one as a
// `rename` redirect; run it in one transaction (SDD 01 §7.1). Redirects that
// already point at id keep doing so, so chains never exceed one hop. Renaming
// back to a slug that redirects to this same row removes that redirect (I-1).
func RenameSlug(ctx context.Context, ex database.Executor, scope Scope, namespace, id, newSlug string) error {
	table, col, err := scope.table()
	if err != nil {
		return err
	}

	var old string
	if err := ex.QueryRowContext(
		ctx,
		fmt.Sprintf(
			`SELECT slug FROM %s WHERE id = $1 AND %s = $2 FOR UPDATE`,
			table,
			col,
		),
		id,
		namespace,
	).Scan(&old); err != nil {
		return fmt.Errorf("load %s %s: %w", scope, id, err)
	}

	if old == newSlug {
		return nil
	}

	var otherTarget bool
	if err := ex.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM catalog_redirects
WHERE scope = $1 AND namespace = $2 AND old_slug = $3 AND target_id <> $4)`,
		string(scope), namespace, newSlug, id).Scan(&otherTarget); err != nil {
		return err
	}

	if otherTarget || reserved[newSlug] {
		return ErrSlugReserved
	}

	if _, err := ex.ExecContext(
		ctx,
		`DELETE FROM catalog_redirects WHERE scope = $1 AND namespace = $2 AND old_slug = $3`,
		string(scope),
		namespace,
		newSlug,
	); err != nil {
		return err
	}

	if _, err := ex.ExecContext(
		ctx,
		fmt.Sprintf(`UPDATE %s SET slug = $1, updated_at = now() WHERE id = $2`, table),
		newSlug,
		id,
	); err != nil {
		return err
	}

	_, err = ex.ExecContext(ctx, `INSERT INTO catalog_redirects (scope, namespace, old_slug, target_id, reason)
VALUES ($1, $2, $3, $4, 'rename')`, string(scope), namespace, old, id)

	return err
}

// RecordMerge leaves the id and slug of source as one `merge` redirect to
// targetID and repoints every redirect that targeted source, so chains stay
// one hop (I-2). The merge service (P1.B.3.2) calls it in its transaction
// before deleting the source row.
func RecordMerge(
	ctx context.Context,
	ex database.Executor,
	scope Scope,
	namespace string,
	source Target,
	targetID string,
) error {
	if _, err := ex.ExecContext(ctx, `UPDATE catalog_redirects SET target_id = $1 WHERE scope = $2 AND target_id = $3`,
		targetID, string(scope), source.ID); err != nil {
		return err
	}

	_, err := ex.ExecContext(ctx, `INSERT INTO catalog_redirects (scope, namespace, old_slug, old_id, target_id, reason)
VALUES ($1, $2, $3, $4, $5, 'merge')`, string(scope), namespace, source.Slug, source.ID, targetID)

	return err
}

// ResolveSlug returns the live row an old slug redirects to.
func ResolveSlug(ctx context.Context, ex database.Executor, scope Scope, namespace, slug string) (Target, bool, error) {
	return resolve(ctx, ex, scope, "old_slug", namespace, slug)
}

// ResolveID returns the live row a merged-away id redirects to.
func ResolveID(ctx context.Context, ex database.Executor, scope Scope, namespace, id string) (Target, bool, error) {
	return resolve(ctx, ex, scope, "old_id", namespace, id)
}

func resolve(
	ctx context.Context,
	ex database.Executor,
	scope Scope,
	keyCol, namespace, key string,
) (Target, bool, error) {
	table, _, err := scope.table()
	if err != nil {
		return Target{}, false, err
	}

	var t Target

	q := fmt.Sprintf(`SELECT t.id::text, t.slug FROM catalog_redirects r JOIN %s t ON t.id = r.target_id
WHERE r.scope = $1 AND r.namespace = $2 AND r.%s = $3`, table, keyCol)

	err = ex.QueryRowContext(ctx, q, string(scope), namespace, key).Scan(&t.ID, &t.Slug)
	if errors.Is(err, sql.ErrNoRows) {
		return Target{}, false, nil
	}

	return t, err == nil, err
}
