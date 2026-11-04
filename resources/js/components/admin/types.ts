/** Props shared by every React admin screen: the panel's one-shot flash messages (app/admin/react.go). */
export interface Flash {
	success: string;
	warning: string;
}

/** One ingestion run (app/admin/catalog_handlers.go runRow). Go sends null for an empty list. */
export interface RunRow {
	id: number;
	job: string;
	source: string;
	status: string;
	error: string;
	started: string;
	finished: string;
	seen: number;
	changed: number;
	failed: number;
	rollbackable: boolean;
}

export interface IngestStatus {
	failures: { source: string; total: number; parked: number }[] | null;
	asserts: { code: string; severity: string; observed: string }[] | null;
	oldestTMDB: string;
	coverage: { type: string; day: string; sampled: number; missing: number }[] | null;
	stages: { job: string; source: string }[] | null;
	modes: string[] | null;
	canRun: boolean;
	canRollback: boolean;
}

export interface SourceRow {
	source: string;
	records: number;
	pending: number;
	lastFetched: string;
	lastChecked: string;
}
