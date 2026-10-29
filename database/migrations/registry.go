// Package migrations holds the schema migrations.
package migrations

import "github.com/AlexGalhardo/vukem/schema"

// All lists the migrations in order; vukem make:migration appends here.
func All() []schema.Migration {
	return []schema.Migration{
		m20261029100000CreateDataSourcesTable(),
		m20261029100001CreateItemsTable(),
		m20261029100002CreateItemTranslationsTable(),
		m20261029100003CreateItemAltTitlesTable(),
	}
}
