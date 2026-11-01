import { render, screen } from "@testing-library/react";
import { expect, test, vi } from "vitest";

import DeleteAccount from "./DeleteAccount";

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	usePage: () => ({ props: { locale: "pt-BR" } }),
	Link: ({ href, children }: { href: string; children: React.ReactNode }) => (
		<a href={href}>{children}</a>
	),
}));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

// Play's deletion URL is opened without the app; a pt-BR visitor must stay in pt-BR (P5.C.2.1).
test("links keep the visitor's locale prefix", () => {
	render(<DeleteAccount />);

	expect(screen.getByText("help.deleteAccount.cta").getAttribute("href")).toBe(
		"/pt-br/profile#delete",
	);
	expect(screen.getByText("help.deleteAccount.login").getAttribute("href")).toBe("/pt-br/login");
	expect(screen.getByText("help.deleteAccount.grace")).toBeTruthy();
});
