import { act, cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import DocumentPage from "./DocumentPage";

vi.mock("./Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ props: { locale: "en" } }),
}));

type Callback = (entries: Array<Partial<IntersectionObserverEntry>>) => void;
let observed: Callback = () => {};

afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
});

// The contents list comes first in the DOM, so it stacks above the text below lg and sits on the left above lg.
test("the contents list is the first element and marks the section in view", () => {
	vi.stubGlobal(
		"IntersectionObserver",
		class {
			constructor(callback: Callback) {
				observed = callback;
			}
			observe() {}
			disconnect() {}
		},
	);
	const { container } = render(
		<DocumentPage
			title="Terms"
			description="d"
			intro="i"
			contentsLabel="Contents"
			sections={[
				{ id: "one", heading: "One", body: <p>a</p> },
				{ id: "two", heading: "Two", body: <p>b</p> },
			]}
		/>,
	);
	expect(container.firstElementChild?.firstElementChild?.tagName).toBe("ASIDE");

	act(() => observed([{ isIntersecting: true, target: { id: "two" } as Element }]));
	expect(screen.getByRole("link", { name: "Two" }).getAttribute("aria-current")).toBe("location");
	expect(screen.getByRole("link", { name: "One" }).getAttribute("aria-current")).toBeNull();
});
