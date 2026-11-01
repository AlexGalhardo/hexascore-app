import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import { activeCount } from "../catalog/Filters";
import ScoreBadge, { formatScore, scoreBand } from "./ScoreBadge";

vi.mock("@inertiajs/react", () => ({ usePage: () => ({ props: { locale: "pt-BR" } }) }));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(cleanup);

test("bands follow the 5.0 and 7.5 thresholds", () => {
	expect(scoreBand(null)).toBe("none");
	expect(scoreBand(4.9)).toBe("low");
	expect(scoreBand(5)).toBe("mid");
	expect(scoreBand(7.4)).toBe("mid");
	expect(scoreBand(7.5)).toBe("high");
});

test("scores print one decimal in the locale, or a dash", () => {
	expect(formatScore(8, "en")).toBe("8.0");
	expect(formatScore(8.25, "pt-BR")).toBe("8,3");
	expect(formatScore(null, "en")).toBe("–");
});

test("the badge prints the number on the band colour", () => {
	render(<ScoreBadge score={9.2} />);
	const badge = screen.getByRole("img");
	expect(badge.textContent).toBe("9,2");
	expect(badge.className).toContain("bg-score-high");
	expect(badge.className).toContain("text-on-score");
});

test("activeCount ignores sort, page size and empty filters", () => {
	expect(
		activeCount({ genre: "", platform: "", year: 0, minScore: null, sort: "-score", per: 50 }),
	).toBe(0);
	expect(
		activeCount({ genre: "rpg", platform: "pc", year: 2020, minScore: 7, sort: "", per: 25 }),
	).toBe(4);
});
