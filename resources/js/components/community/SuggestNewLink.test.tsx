import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import SuggestNewLink from "./SuggestNewLink";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
}));

afterEach(cleanup);

test("links to the create form of the domain", () => {
	render(<SuggestNewLink domain="game" />);
	expect(screen.getByRole("link").getAttribute("href")).toBe("/suggest/new/game");
});
