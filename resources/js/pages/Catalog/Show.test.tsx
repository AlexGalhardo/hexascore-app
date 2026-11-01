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
vi.mock("../../components/Seo", () => ({ default: () => null, truncate: (s: string) => s }));
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
	expect(screen.getByText(/item.band.high/)).toBeTruthy();
	expect(screen.getByText(/item.scoreFrom.*1,482/)).toBeTruthy();
	expect(screen.getAllByRole("link", { name: "MMORPG" })[0].getAttribute("href")).toBe(
		"/genres/mmorpg",
	);
	expect(screen.getByText("itemTypes.game · 2004 · Blizzard Entertainment")).toBeTruthy();
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
