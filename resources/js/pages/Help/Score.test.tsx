import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import Score from "./Score";

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: () => null,
	usePage: () => ({ props: { locale: "en", appName: "HexaScore" } }),
}));

afterEach(cleanup);

test("uses the document layout: anchored sections listed in the contents", () => {
	render(<Score categoryKeys={{ game: ["story"] }} />);
	expect(screen.getByRole("heading", { level: 1, name: "help.score.title" })).toBeTruthy();
	expect(screen.getByRole("region", { name: "help.score.categoriesTitle" }).id).toBe(
		"categories",
	);
	expect(screen.getByRole("link", { name: "help.score.formulaTitle" }).getAttribute("href")).toBe(
		"#formula",
	);
	expect(screen.getByText("categories.game.story")).toBeTruthy();
});
