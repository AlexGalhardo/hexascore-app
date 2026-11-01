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
