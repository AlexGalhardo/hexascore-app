import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import { cleanup, render, screen } from "../../lib/testing";
import Index from "./Index";

afterEach(cleanup);

const props = vi.hoisted(() => ({ locale: "pt-BR", appName: "HexaScore" }));

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({
		href,
		children,
		className,
	}: {
		href: string;
		children: ReactNode;
		className?: string;
	}) => (
		<a href={href} className={className}>
			{children}
		</a>
	),
	usePage: () => ({ url: "/pt-br/about", props }),
}));

// Owner round 2026-11-01 item 25.
test("the about page reads at legal width and links categories, scores, docs, source and contact", () => {
	const { container } = render(<Index />);
	expect(container.querySelector("article")?.className).toContain("max-w-[45rem]");
	expect(screen.getByRole("heading", { level: 1 }).textContent).toBe("about.title");
	expect(screen.getAllByRole("heading", { level: 2 }).map((h) => h.textContent)).toEqual([
		"about.categoriesTitle",
		"about.scoresTitle",
		"about.communityTitle",
		"about.apiTitle",
		"about.openSourceTitle",
		"about.contactTitle",
	]);
	const href = (name: string) => screen.getByRole("link", { name }).getAttribute("href");
	expect(href("catalog.title.game")).toBe("/pt-br/games");
	expect(
		screen
			.getByRole("link", { name: "catalog.title.book" })
			.querySelector("[data-category-icon]"),
	).toBeTruthy();
	expect(href("about.scoresLink")).toBe("/pt-br/how-score-works");
	expect(href("about.apiLink")).toBe("/docs");
	expect(href("about.contactLink")).toBe("/pt-br/contact");
	const source = screen.getByRole("link", { name: "about.openSourceLink" });
	expect(source.getAttribute("href")).toBe("https://github.com/AlexGalhardo/hexascore-app");
	expect(source.getAttribute("target")).toBe("_blank");
	expect(source.getAttribute("rel")).toBe("noopener noreferrer");
});
