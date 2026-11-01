import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, expect, test, vi } from "vitest";

import Page from "./Page";

const dict: Record<string, string> = {
	"terms.title": "Terms",
	"terms.intro": "Intro",
	"terms.sections.0.heading": "1. The service",
	"terms.sections.0.body.0": "We run a database.",
	"terms.sections.1.heading": "2. Your account",
	"terms.sections.1.body.0": "Keep it safe.",
	"terms.otherTitle": "Privacy policy",
};

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => dict[key] ?? key }));
vi.mock("../../components/Seo", () => ({ default: () => null }));
vi.mock("@inertiajs/react", () => ({
	Link: ({ href, children }: { href: string; children: ReactNode }) => (
		<a href={href}>{children}</a>
	),
	usePage: () => ({ props: { locale: "en", appName: "HexaScore" } }),
}));

afterEach(cleanup);

test("numbered sections get anchors listed in the contents, plus a link to the other document", () => {
	render(<Page slug="terms" updatedAt="2026-10-30T00:00:00Z" />);
	expect(screen.getByRole("region", { name: "2. Your account" }).id).toBe("section-2");
	expect(screen.getByRole("link", { name: "1. The service" }).getAttribute("href")).toBe(
		"#section-1",
	);
	expect(screen.getByRole("link", { name: "Privacy policy" }).getAttribute("href")).toBe(
		"/policy",
	);
	expect(screen.getByText("October 30, 2026")).toBeTruthy();
});
