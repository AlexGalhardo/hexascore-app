export type PageSlot = number | "gap";

/** Compact numbered pagination: first, last and current ± 1, with gaps. */
export function pageWindow(current: number, count: number): PageSlot[] {
	const pages = new Set([1, count, current - 1, current, current + 1]);
	const sorted = [...pages].filter((p) => p >= 1 && p <= count).sort((a, b) => a - b);
	const slots: PageSlot[] = [];
	for (const p of sorted) {
		const prev = slots.at(-1);
		// A single hidden page is shown instead of a gap, which would take the same room.
		if (typeof prev === "number" && p - prev === 2) slots.push(prev + 1);
		else if (typeof prev === "number" && p - prev > 2) slots.push("gap");
		slots.push(p);
	}
	return slots;
}

/** Short column header: "Gameplay" → "Game", "Qualidade sonora" → "QS". */
export function abbreviateLabel(label: string): string {
	const words = label.split(/[\s/]+/).filter(Boolean);
	return words.length > 1
		? words.map((w) => w.charAt(0).toUpperCase()).join("")
		: label.slice(0, 4);
}
