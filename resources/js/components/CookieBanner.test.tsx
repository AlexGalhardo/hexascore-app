import userEvent from "@testing-library/user-event";
import { act } from "react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { openCookieSettings } from "../lib/preferences";
import { cleanup, render, screen } from "../lib/testing";

import CookieBanner from "./CookieBanner";

const props = vi.hoisted(() => ({ consent: undefined as undefined | "all" | "essential" }));
vi.mock("@inertiajs/react", () => ({ usePage: () => ({ props }) }));
vi.mock("../lib/useT", () => ({ useT: () => (key: string) => key }));

const consent = vi.fn();

function clearConsentCookie(): void {
	// biome-ignore lint/suspicious/noDocumentCookie: test reset of the preference cookie
	document.cookie = "consent=; path=/; max-age=0";
}

beforeEach(() => {
	clearConsentCookie();
	props.consent = undefined;
	consent.mockReset();
	window.vukemAnalytics = { consent };
});

afterEach(() => {
	cleanup();
	delete window.vukemAnalytics;
});

test("no choice: banner shown, GA consent untouched", () => {
	render(<CookieBanner />);
	expect(screen.getByRole("region", { name: "cookies.label" })).toBeTruthy();
	expect(consent).not.toHaveBeenCalled();
});

test("accept grants GA consent and stores a first-party cookie", async () => {
	render(<CookieBanner />);
	await userEvent.click(screen.getByRole("button", { name: "cookies.accept" }));
	expect(consent).toHaveBeenLastCalledWith(true);
	expect(document.cookie).toContain("consent=all");
	expect(screen.queryByRole("region", { name: "cookies.label" })).toBeNull();
});

test("reject denies GA consent", async () => {
	render(<CookieBanner />);
	await userEvent.click(screen.getByRole("button", { name: "cookies.decline" }));
	expect(consent).toHaveBeenLastCalledWith(false);
	expect(document.cookie).toContain("consent=essential");
});

test("a stored choice is replayed on load and the banner stays closed", () => {
	props.consent = "all";
	// biome-ignore lint/suspicious/noDocumentCookie: test fixture of the preference cookie
	document.cookie = "consent=all; path=/";
	render(<CookieBanner />);
	expect(consent).toHaveBeenCalledWith(true);
	expect(screen.queryByRole("region", { name: "cookies.label" })).toBeNull();
});

test("cookie settings reopens the banner with focus on its first button", async () => {
	props.consent = "essential";
	render(<CookieBanner />);
	act(() => openCookieSettings());
	const reject = screen.getByRole("button", { name: "cookies.decline" });
	expect(document.activeElement).toBe(reject);
	await userEvent.click(screen.getByRole("button", { name: "cookies.accept" }));
	expect(consent).toHaveBeenLastCalledWith(true);
});

test("without GA configured the choice still works", async () => {
	delete window.vukemAnalytics;
	render(<CookieBanner />);
	await userEvent.click(screen.getByRole("button", { name: "cookies.accept" }));
	expect(document.cookie).toContain("consent=all");
});
