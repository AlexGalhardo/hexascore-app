/** The snapshot served by app/admin/overview.go (OverviewData). */
export interface OverviewRun {
	id: number;
	job: string;
	source: string;
	mode: string;
	status: string;
	seen: number;
	changed: number;
	failed: number;
	startedAt: string | null;
	durationSeconds: number | null;
}

export interface OverviewData {
	generatedAt: string;
	counts: {
		items: number;
		users: number;
		openReports: number;
		runningRuns: number;
		failedRuns24h: number;
		parkedFailures: number;
	};
	worker: { status: "ok" | "stale" | "never"; ageSeconds: number | null };
	runs: OverviewRun[];
}

export const OVERVIEW_URL = "/admin/overview/data";

/** 83 -> "1m 23s"; the unit set stays short because the table cell is narrow. */
export function formatDuration(totalSeconds: number): string {
	if (totalSeconds < 60) return `${totalSeconds}s`;
	const minutes = Math.floor(totalSeconds / 60);
	const seconds = totalSeconds % 60;
	if (minutes < 60) return seconds === 0 ? `${minutes}m` : `${minutes}m ${seconds}s`;
	const hours = Math.floor(minutes / 60);
	const rest = minutes % 60;
	return rest === 0 ? `${hours}h` : `${hours}h ${rest}m`;
}

export type RunVariant = "success" | "warning" | "destructive" | "secondary";

/** The badge colour of a run status; running and queued stay neutral. */
export function runVariant(status: string): RunVariant {
	switch (status) {
		case "succeeded":
			return "success";
		case "partial":
			return "warning";
		case "failed":
			return "destructive";
		default:
			return "secondary";
	}
}

export async function fetchOverview(): Promise<OverviewData> {
	const res = await fetch(OVERVIEW_URL, {
		headers: { Accept: "application/json" },
		credentials: "same-origin",
	});
	if (!res.ok) throw new Error(`overview ${res.status}`);
	return (await res.json()) as OverviewData;
}
