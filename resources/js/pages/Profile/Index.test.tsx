import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import Profile from "./Index";

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	Link: ({ children, href }: { children: React.ReactNode; href: string }) => (
		<a href={href}>{children}</a>
	),
	router: { delete: () => undefined, post: () => undefined, reload: () => undefined },
	usePage: () => ({ props: { auth: { user: null } } }),
	useForm: <T extends object>(data: T) => ({
		data,
		errors: {},
		processing: false,
		setData: () => undefined,
		patch: () => undefined,
		put: () => undefined,
		post: () => undefined,
		reset: () => undefined,
	}),
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(cleanup);

const baseProps = {
	details: { name: "Ana", email: "ana@example.com", locale: "en" },
	passwordSet: true,
	twoFactor: { enabled: false },
	sessions: [],
	privacy: { profilePrivate: false, personalizationOptOut: false },
	notifications: { communityMailOptOut: false },
};

test("the profile forms sit in a narrow centred column", () => {
	render(<Profile {...baseProps} />);

	expect(screen.getByTestId("profile-page").className).toContain("max-w-2xl");
	expect(screen.getByTestId("profile-page").className).toContain("mx-auto");
});
