import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import ErrorPage from "./Error";

vi.mock("../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ props: { locale: "pt-BR" } }),
}));

afterEach(cleanup);

test("404 offers home and a search form that submits q to the localized search page", () => {
	render(<ErrorPage status={404} message="Not Found" />);
	expect(screen.getByRole("heading", { level: 1, name: "errors.notFoundTitle" })).toBeTruthy();
	expect(screen.getByRole("link", { name: "errors.home" }).getAttribute("href")).toBe("/pt-br/");
	const form = screen.getByRole("searchbox").closest("form") as HTMLFormElement;
	expect(form.getAttribute("action")).toBe("/pt-br/search");
	expect(screen.getByRole("searchbox", { name: "search.label" }).getAttribute("name")).toBe("q");
	expect(screen.queryByText("errors.retry")).toBeNull();
});

test("500 offers a retry and home, without search", () => {
	render(<ErrorPage status={500} message="Internal Server Error" />);
	expect(screen.getByRole("heading", { level: 1, name: "errors.serverTitle" })).toBeTruthy();
	expect(screen.getByRole("button", { name: "errors.retry" })).toBeTruthy();
	expect(screen.queryByRole("searchbox")).toBeNull();
});
