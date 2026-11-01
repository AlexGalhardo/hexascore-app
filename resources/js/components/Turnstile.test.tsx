import { act } from "react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { cleanup, render, screen } from "../lib/testing";

import { CHALLENGE_TOKEN_HEADER, useTurnstile } from "./Turnstile";

const props = vi.hoisted(() => ({ turnstileSiteKey: "" }));
vi.mock("@inertiajs/react", () => ({ usePage: () => ({ props }) }));

type Opts = { sitekey: string; action: string; callback: (t: string) => void };
const rendered: Opts[] = [];
const remove = vi.fn();
let latest: ReturnType<typeof useTurnstile> | null = null;

function Probe({ error }: { error?: string }) {
	const ts = useTurnstile("login");
	latest = ts;
	return <>{ts.render(error)}</>;
}

beforeEach(() => {
	rendered.length = 0;
	remove.mockReset();
	props.turnstileSiteKey = "";
	window.turnstile = {
		render: (_el, opts) => {
			rendered.push(opts as unknown as Opts);
			return `w${rendered.length}`;
		},
		remove,
	};
});

afterEach(() => {
	cleanup();
	latest = null;
	delete window.turnstile;
});

test("off without a site key: no widget and no header", () => {
	render(<Probe error="x" />);
	expect(screen.queryByTestId("turnstile")).toBeNull();
	expect(latest?.headers()).toEqual({});
	expect(rendered).toHaveLength(0);
});

test("renders the widget for the action and sends the solved token", async () => {
	props.turnstileSiteKey = "1x00000000000000000000AA";
	render(<Probe />);
	await act(async () => {});
	expect(rendered).toHaveLength(1);
	expect(rendered[0]).toMatchObject({ sitekey: "1x00000000000000000000AA", action: "login" });

	act(() => rendered[0]?.callback("tok-1"));
	expect(latest?.headers()).toEqual({ [CHALLENGE_TOKEN_HEADER]: "tok-1" });
});

test("reset drops the used token and renders a fresh widget", async () => {
	props.turnstileSiteKey = "1x00000000000000000000AA";
	render(<Probe />);
	await act(async () => {});
	act(() => rendered[0]?.callback("tok-1"));

	await act(async () => latest?.reset());
	expect(latest?.headers()).toEqual({});
	expect(remove).toHaveBeenCalledWith("w1");
	expect(rendered).toHaveLength(2);
});

test("shows the server's challenge error as an alert", () => {
	props.turnstileSiteKey = "1x00000000000000000000AA";
	render(<Probe error="Complete the check." />);
	expect(screen.getByRole("alert").textContent).toBe("Complete the check.");
});
