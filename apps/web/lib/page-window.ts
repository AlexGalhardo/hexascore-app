export type PageSlot = number | "gap";

/** Compact numbered pagination: first, last and current ± 1, with gaps (e.g. 1 … 4 5 6 … 100). */
export function pageWindow(current: number, count: number): PageSlot[] {
	const pages = new Set([1, count, current - 1, current, current + 1]);
	const sorted = [...pages].filter((p) => p >= 1 && p <= count).sort((a, b) => a - b);
	const slots: PageSlot[] = [];
	for (const p of sorted) {
		const prev = slots.at(-1);
		// a single hidden page is shown instead of a gap, which would take the same room
		if (typeof prev === "number" && p - prev === 2) slots.push(prev + 1);
		else if (typeof prev === "number" && p - prev > 2) slots.push("gap");
		slots.push(p);
	}
	return slots;
}
