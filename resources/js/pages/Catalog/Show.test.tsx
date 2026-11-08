import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";
import { cleanup, render, screen } from "../../lib/testing";

import Show from "./Show";

afterEach(cleanup);

let auth: { user: { id: string } } | null = null;

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));
const seoImages = vi.hoisted((): (string | undefined)[] => []);
vi.mock("../../components/Seo", () => ({
	default: ({ image }: { image?: string }) => {
		seoImages.push(image);
		return null;
	},
	truncate: (s: string) => s,
}));
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
	usePage: () => ({ props: { locale: "en", coversBase: "/c/", auth } }),
	router: { put: vi.fn(), delete: vi.fn() },
}));

const props = {
	item: {
		id: "i1",
		type: "game" as const,
		slug: "world-of-warcraft",
		title: "World of Warcraft",
		releaseDate: "2004-11-23",
	},
	ratingSummary: {
		score: 8.1,
		scoreByCategory: {
			gameplay: 8.6,
			fun: 8.4,
			sound: 8.8,
			technical: 7.2,
			visual: 7.9,
			storytelling: 8,
		},
		ratingCount: 1482,
		confidence: null,
		unusualActivity: false,
		ratingsLimitedUntil: null,
	},
	entities: {
		developer: [
			{ id: "d1", kind: "organization", slug: "blizzard", name: "Blizzard Entertainment" },
		],
		genre: [{ id: "g1", kind: "genre", slug: "mmorpg", name: "MMORPG" }],
		platform: [{ id: "p1", kind: "platform", slug: "macos", name: "macOS" }],
	},
	editions: null,
	editionCount: 0,
	timeToBeat: null,
};

test("logged out: score block with explanation, chips, disabled stars and a log in call", () => {
	auth = null;
	render(<Show {...props} />);
	expect(screen.getByRole("heading", { level: 1 }).textContent).toBe("World of Warcraft");
	expect(screen.getByText(/item.verdict.excellent/)).toBeTruthy();
	expect(screen.getByText(/item.scoreFrom.*1,482/)).toBeTruthy();
	expect(screen.getAllByRole("link", { name: "MMORPG" })[0].getAttribute("href")).toBe(
		"/games?genre=mmorpg&per=40&sort=-popularity",
	);
	expect(screen.getByRole("link", { name: "Blizzard Entertainment" }).className).toContain(
		"text-blue-700",
	);
	expect(screen.queryByText(/itemTypes.game · 2004/)).toBeNull();
	expect(screen.queryByText(/item.attribution/)).toBeNull();
	expect(screen.getAllByRole("radiogroup")).toHaveLength(6);
	expect(screen.getAllByRole("radio").every((r) => (r as HTMLInputElement).disabled)).toBe(true);
	expect(screen.getByRole("link", { name: "item.loginToRate" }).getAttribute("href")).toBe(
		"/login",
	);
	expect(screen.queryByRole("button", { name: /ratings.save/ })).toBeNull();
	expect(screen.queryByRole("group", { name: "social.status.label" })).toBeNull();
});

test("logged in: enabled stars, overall preview, save and status picker", () => {
	auth = { user: { id: "u1" } };
	render(
		<Show
			{...props}
			me={{ rating: null, status: "in_progress", timeToBeat: null }}
			myLists={[]}
		/>,
	);
	expect(screen.getAllByRole("radio").some((r) => (r as HTMLInputElement).disabled)).toBe(false);
	expect(
		(screen.getByRole("button", { name: "ratings.save" }) as HTMLButtonElement).disabled,
	).toBe(true);
	expect(screen.getByRole("group", { name: "social.status.label" })).toBeTruthy();
	expect(
		screen
			.getByRole("button", { name: "social.status.game.in_progress" })
			.getAttribute("aria-pressed"),
	).toBe("true");
	expect(screen.queryByRole("link", { name: "item.loginToRate" })).toBeNull();
});

test("unreleased item explains the missing score and blocks guest and member rating controls", () => {
	auth = null;
	const guest = render(<Show {...props} unreleased />);
	expect(screen.getByText("item.unreleased")).toBeTruthy();
	expect(screen.getByTestId("score-block").textContent).toContain("item.unreleasedHint");
	expect(screen.queryByText(/item.verdict.excellent/)).toBeNull();
	expect(
		(screen.getByRole("button", { name: "item.loginToRate" }) as HTMLButtonElement).disabled,
	).toBe(true);
	expect(screen.queryByRole("link", { name: "item.loginToRate" })).toBeNull();
	expect(screen.getAllByRole("radio").every((r) => (r as HTMLInputElement).disabled)).toBe(true);
	guest.unmount();

	auth = { user: { id: "u1" } };
	render(<Show {...props} unreleased me={{ rating: null, status: null, timeToBeat: null }} />);
	expect(screen.getAllByRole("radio").every((r) => (r as HTMLInputElement).disabled)).toBe(true);
	expect(
		(screen.getByRole("button", { name: "ratings.save" }) as HTMLButtonElement).disabled,
	).toBe(true);
});

test("three columns: actions under the cover, title and synopsis before the score card and rating form", () => {
	auth = null;
	render(
		<Show
			{...props}
			item={{ ...props.item, synopsis: { text: "An online world.", licence: "CC BY-SA" } }}
		/>,
	);

	const actions = screen.getByTestId("item-col-actions");
	const names = [...actions.querySelectorAll("a, button")]
		.filter((el) => !el.closest("section"))
		.map((el) => el.textContent);
	expect(names).toEqual(["item.seeUserScores", "lists.addToList"]);
	for (const el of [...actions.querySelectorAll("a, button")].filter(
		(e) => !e.closest("section"),
	)) {
		expect(el.querySelector("svg[aria-hidden='true']")).not.toBeNull();
	}
	expect(actions.firstElementChild?.matches("img, [data-testid=cover-placeholder]")).toBe(true);
	expect(actions.querySelector("#details-heading")).not.toBeNull();

	// No genre or platform buttons under the title (owner 2026-11-02); Details still links them.
	const main = screen.getByTestId("item-col-main");
	expect(screen.queryByTestId("item-chips")).toBeNull();
	expect(main.querySelector("header")?.nextElementSibling?.getAttribute("aria-labelledby")).toBe(
		"synopsis-heading",
	);
	// Phone width (OQ-15, 2026-11-02): title and synopsis come right after Details, before the score card and rating.
	expect(main.className).not.toContain("max-sm:order-last");
	const rating = screen.getByTestId("item-col-rating");
	expect(actions.compareDocumentPosition(main) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
	expect(main.compareDocumentPosition(rating) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
	const score = screen.getByTestId("score-block");
	const form = document.getElementById("your-rating") as HTMLElement;
	expect(rating.contains(score) && rating.contains(form)).toBe(true);
	expect(score.compareDocumentPosition(form) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
});

test("the item page blocks mouse text selection but keeps fields selectable", () => {
	auth = null;
	render(<Show {...props} />);

	const page = screen.getByTestId("item-page");
	expect(page.className).toContain("select-none");
	expect(page.className).toContain("[&_input]:select-text");
	expect(page.className).toContain("[&_textarea]:select-text");
});

test("synopsis shows no licence line, only the source link; no recommendations block", () => {
	auth = null;
	render(
		<Show
			{...props}
			item={{
				...props.item,
				synopsis: {
					text: "An online world.",
					licence: "CC BY-SA",
					sourceUrl: "https://en.wikipedia.org/wiki/WoW",
				},
			}}
		/>,
	);
	expect(screen.queryByText(/item.license/)).toBeNull();
	expect(screen.getByRole("link", { name: "item.source" }).getAttribute("href")).toBe(
		"https://en.wikipedia.org/wiki/WoW",
	);
	expect(screen.queryByText(/item.moreLike/)).toBeNull();
});

test("details: each fact is a heading with its values below, one linked genre per line", () => {
	auth = null;
	render(
		<Show
			{...props}
			entities={{
				...props.entities,
				genre: [
					{ id: "g1", kind: "genre", slug: "mmorpg", name: "MMORPG" },
					{ id: "g2", kind: "genre", slug: "rpg", name: "RPG" },
				],
			}}
		/>,
	);
	const details = screen.getByTestId("item-col-actions").querySelector("dl") as HTMLElement;
	const genres = [...details.querySelectorAll("dt")].find(
		(d) => d.textContent === "item.roles.genre",
	);
	const items = genres?.nextElementSibling?.querySelectorAll("li a");
	expect([...(items ?? [])].map((a) => a.textContent)).toEqual(["MMORPG", "RPG"]);
});

test("headings of the key sections are bold", () => {
	auth = null;
	render(<Show {...props} />);
	for (const id of [
		"details-heading",
		"synopsis-heading",
		"averages-heading",
		"rating-heading",
	]) {
		expect(document.getElementById(id)?.className).toContain("font-bold");
	}
	expect(screen.getByRole("heading", { level: 1 }).className).toContain("font-bold");
});

test("share image: the cover when no share card is set, the share card when it is", () => {
	const cover = { thumb: "/covers/games/wow-t.webp", full: "/covers/games/wow.webp" };
	seoImages.length = 0;
	render(<Show {...props} cover={cover} />);
	expect(seoImages.at(-1)).toBe(cover.full);
	cleanup();

	seoImages.length = 0;
	render(<Show {...props} cover={cover} ogImage="/og/items/i1.png" />);
	expect(seoImages.at(-1)).toBe("/og/items/i1.png");
});
