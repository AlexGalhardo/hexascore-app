import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import TwoFactorDialog from "./TwoFactorDialog";

const post = vi.fn();
const transform = vi.fn();

vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("@inertiajs/react", () => ({
	useForm: () => ({
		data: { code: "", recovery_code: "" },
		errors: {},
		processing: false,
		post,
		transform,
		setData: vi.fn(),
		reset: vi.fn(),
		clearErrors: vi.fn(),
	}),
}));

afterEach(() => {
	cleanup();
	post.mockClear();
	transform.mockClear();
});

const boxes = () => screen.getAllByRole("textbox", { name: /auth\.login\.code \d/ });

test("six one-digit boxes submit the code on their own once all six are typed", () => {
	render(<TwoFactorDialog open />);
	expect(boxes()).toHaveLength(6);

	"12345".split("").forEach((d, i) => {
		fireEvent.change(boxes()[i], { target: { value: d } });
	});
	expect(post).not.toHaveBeenCalled();

	fireEvent.change(boxes()[5], { target: { value: "6" } });
	expect(post).toHaveBeenCalledWith("/two-factor-challenge", expect.anything());
	expect(transform.mock.calls.at(-1)?.[0]()).toEqual({ code: "123456", recovery_code: "" });
});

test("a pasted code fills every box and submits", () => {
	render(<TwoFactorDialog open />);
	fireEvent.paste(boxes()[0], { clipboardData: { getData: () => " 987 654 " } });

	expect(boxes().map((b) => (b as HTMLInputElement).value)).toEqual([
		"9",
		"8",
		"7",
		"6",
		"5",
		"4",
	]);
	expect(transform.mock.calls.at(-1)?.[0]()).toEqual({ code: "987654", recovery_code: "" });
});

test("letters are ignored and nothing is sent while a box is empty", () => {
	render(<TwoFactorDialog open />);
	fireEvent.change(boxes()[0], { target: { value: "a" } });

	expect((boxes()[0] as HTMLInputElement).value).toBe("");
	expect(post).not.toHaveBeenCalled();
});
