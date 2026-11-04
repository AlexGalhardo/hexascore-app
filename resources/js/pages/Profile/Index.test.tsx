import { cleanup, fireEvent, render, screen } from "@testing-library/react";
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

test("a section menu lists every section and each link points at its section", () => {
	render(<Profile {...baseProps} />);

	const nav = screen.getByRole("navigation", { name: "account.nav" });
	const ids = [
		"details",
		"password",
		"two-factor",
		"passkeys",
		"sessions",
		"privacy",
		"notifications",
		"export",
		"delete",
	];
	for (const id of ids) {
		const link = nav.querySelector(`a[href="#${id}"]`);
		expect(link, id).not.toBeNull();
		expect(document.getElementById(id), id).not.toBeNull();
	}
	expect(nav.querySelectorAll("a")).toHaveLength(ids.length);
	expect(nav.querySelector('a[aria-current="true"]')?.getAttribute("href")).toBe("#details");
});

test("clicking a menu link marks that section as current", () => {
	render(<Profile {...baseProps} />);

	const nav = screen.getByRole("navigation", { name: "account.nav" });
	fireEvent.click(nav.querySelector('a[href="#privacy"]') as HTMLElement);
	expect(nav.querySelector('a[aria-current="true"]')?.getAttribute("href")).toBe("#privacy");
});

const section = (): HTMLElement => document.getElementById("two-factor") as HTMLElement;

test("two-factor off: the profile section offers to enable it", () => {
	render(<Profile {...baseProps} />);

	expect(section().textContent).toContain("account.twoFactor.off");
	expect(screen.getByRole("button", { name: "account.twoFactor.enable" })).toBeTruthy();
});

test("two-factor pending: QR code, setup key and the code form", () => {
	render(
		<Profile
			{...baseProps}
			twoFactor={{
				enabled: false,
				pending: true,
				secret: "SECRETKEY",
				qr: "data:image/svg+xml;base64,AA",
			}}
		/>,
	);

	expect(screen.getByAltText("account.twoFactor.qrAlt").getAttribute("src")).toBe(
		"data:image/svg+xml;base64,AA",
	);
	expect(section().querySelector("code")?.textContent).toBe("SECRETKEY");
	expect(screen.getByLabelText("account.twoFactor.code").getAttribute("autocomplete")).toBe(
		"one-time-code",
	);
});

test("two-factor on: recovery codes shown once, regenerate and disable ask for the password", () => {
	render(
		<Profile
			{...baseProps}
			twoFactor={{
				enabled: true,
				recoveryCodes: ["aaaa-bbbb", "cccc-dddd"],
				recoveryCodesLeft: 8,
			}}
		/>,
	);

	expect(screen.getByTestId("recovery-codes").textContent).toBe("aaaa-bbbbcccc-dddd");
	expect(section().querySelectorAll("input[type=password][name=password]")).toHaveLength(2);
	expect(screen.getByRole("button", { name: "account.twoFactor.regenerate" })).toBeTruthy();
	expect(screen.getByRole("button", { name: "account.twoFactor.disable" })).toBeTruthy();
});

test("two-factor on, codes already seen: only the count is shown", () => {
	render(
		<Profile
			{...baseProps}
			twoFactor={{ enabled: true, recoveryCodes: null, recoveryCodesLeft: 6 }}
		/>,
	);

	expect(screen.queryByTestId("recovery-codes")).toBeNull();
	expect(section().textContent).toContain("account.twoFactor.recoveryLeft");
});

test("no password yet: the section explains a password is needed and offers no controls", () => {
	render(<Profile {...baseProps} passwordSet={false} />);

	expect(section().querySelector("[role=note]")?.textContent).toBe(
		"account.twoFactor.needPassword",
	);
	expect(screen.queryByRole("button", { name: "account.twoFactor.enable" })).toBeNull();
});
