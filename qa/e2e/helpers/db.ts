// Fixtures come from the DemoCatalog seeder (database/seeders/demo_catalog.go); the suite never writes
// password hashes. Keep these constants in sync with that file.
export const DEMO_PASSWORD = "Demo-password-1";
export const DEMO_ADMIN_EMAIL = "admin@example.test";

// Notifications have no UI trigger a spec can reach cheaply (they come from moderation and suggestion
// decisions), so the spec inserts the row the app itself would write. The Postgres container is the app
// container's sibling in the same compose project; E2E_DB_CONTAINER overrides it.
export function dbContainer(): string {
	return (
		process.env.E2E_DB_CONTAINER ??
		(process.env.E2E_APP_CONTAINER ?? "hexascore-app-1").replace(/-app-/, "-postgres-")
	);
}
