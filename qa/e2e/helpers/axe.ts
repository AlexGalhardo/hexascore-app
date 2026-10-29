import AxeBuilder from "@axe-core/playwright";
import { expect, type Page } from "@playwright/test";

// Serious and critical violations fail the run; minor and moderate are reported by axe but tolerated.
export async function expectNoSeriousA11yViolations(page: Page): Promise<void> {
	const { violations } = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa", "wcag21aa"]).analyze();
	const blocking = violations.filter((v) => v.impact === "serious" || v.impact === "critical");
	expect(blocking.map((v) => `${v.id} (${v.impact}): ${v.nodes.length} node(s)`)).toEqual([]);
}
