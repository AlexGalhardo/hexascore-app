// A17 projected Railway spend (SDD 12 §7.5). No new dependency.
// Programmatic usage is not verified: without USAGE_USD the script only marks a day-20 reminder.

import { appendFileSync } from "node:fs";

export const alertThresholdUsd = 27;
export const alertFromDay = 20;

export function projectedUsd(monthToDateUsd: number, elapsedDays: number, cycleDays: number): number {
	if (elapsedDays <= 0 || cycleDays <= 0) {
		throw new Error("elapsedDays and cycleDays must be positive");
	}

	return (monthToDateUsd * cycleDays) / elapsedDays;
}

// Alerts from day 20 onward when the projection is above US$27. US$27 itself does not alert.
export function shouldAlert(monthToDateUsd: number, elapsedDays: number, cycleDays: number): boolean {
	if (elapsedDays < alertFromDay) {
		return false;
	}

	return projectedUsd(monthToDateUsd, elapsedDays, cycleDays) > alertThresholdUsd;
}

export function daysInUtcMonth(now: Date): number {
	return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 0)).getUTCDate();
}

function writeOutput(alert: boolean, reminder: boolean, path: string): void {
	const file = process.env.GITHUB_OUTPUT;
	if (!file) {
		return;
	}

	appendFileSync(file, `alert=${alert}\nreminder=${reminder}\npath=${path}\n`);
}

function main(): void {
	const usage = process.env.USAGE_USD;
	const elapsed = Number(process.env.ELAPSED_DAYS ?? new Date().getUTCDate());
	const cycle = Number(process.env.CYCLE_DAYS ?? daysInUtcMonth(new Date()));

	if (usage !== undefined && usage !== "") {
		const monthToDate = Number(usage);
		const projected = projectedUsd(monthToDate, elapsed, cycle);
		const alert = shouldAlert(monthToDate, elapsed, cycle);
		console.log(`path=usage projected=${projected} alert=${alert}`);
		writeOutput(alert, false, "usage");

		return;
	}

	const reminder = elapsed === alertFromDay;
	console.log(reminder ? "path=calendar-reminder" : "path=calendar-skip");
	writeOutput(false, reminder, "calendar");
}

if (import.meta.main) {
	main();
}
