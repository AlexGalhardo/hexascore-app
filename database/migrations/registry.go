// Package migrations holds the schema migrations.
package migrations

import "github.com/AlexGalhardo/vukem/schema"

// All lists the migrations in order; vukem make:migration appends here.
func All() []schema.Migration {
	return []schema.Migration{
		m20261029022752CreateUsersTable(),
		m20261029022753CreateAuthTables(),
		m20261029022754CreateSocialAccountsTable(),
		m20261029022755AddTwoFactorColumns(),
		m20261029100000CreateDataSourcesTable(),
		m20261029100001CreateItemsTable(),
		m20261029100002CreateItemTranslationsTable(),
		m20261029100003CreateItemAltTitlesTable(),
		m20261029110000AddProfileColumnsToUsersTable(),
		m20261029110001CreateAccountEventsTable(),
	}
}
