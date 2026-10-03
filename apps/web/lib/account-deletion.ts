// Client-safe helpers (no db imports): used by the delete dialog, the banner and the purge job.
export const DELETION_GRACE_DAYS = 30;
export const DELETION_CONFIRM_SECONDS = 10;
const DAY_MS = 24 * 60 * 60 * 1000;

export function deletionDate(scheduledAt: Date): Date {
	return new Date(scheduledAt.getTime() + DELETION_GRACE_DAYS * DAY_MS);
}

export function purgeCutoff(now: Date): Date {
	return new Date(now.getTime() - DELETION_GRACE_DAYS * DAY_MS);
}

export function secondsLeft(startedAtMs: number, nowMs: number, total = DELETION_CONFIRM_SECONDS): number {
	return Math.max(0, total - Math.floor((nowMs - startedAtMs) / 1000));
}
