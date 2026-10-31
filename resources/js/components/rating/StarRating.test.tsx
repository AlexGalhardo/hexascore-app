import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { afterEach, expect, test, vi } from "vitest";

import StarRating from "./StarRating";

afterEach(cleanup);

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, params?: Record<string, string | number>) =>
		key === "rating.valuetext" ? `${params?.value} of 10 stars` : "Not rated",
}));

function Harness({
	initial = null,
	disabled = false,
}: {
	initial?: number | null;
	disabled?: boolean;
}) {
	const [v, setV] = useState<number | null>(initial);
	return <StarRating label="Gameplay" value={v} onChange={setV} disabled={disabled} />;
}

const slider = () => screen.getByRole("slider", { name: "Gameplay" });

test("exposes slider ARIA with unset and valuetext", () => {
	const { rerender } = render(<StarRating label="Gameplay" value={null} onChange={() => {}} />);
	expect(slider().getAttribute("aria-valuenow")).toBe("0");
	expect(slider().getAttribute("aria-valuetext")).toBe("Not rated");
	rerender(<StarRating label="Gameplay" value={7.5} onChange={() => {}} />);
	expect(slider().getAttribute("aria-valuenow")).toBe("7.5");
	expect(slider().getAttribute("aria-valuetext")).toBe("7.5 of 10 stars");
	expect(slider().getAttribute("aria-valuemax")).toBe("10");
});

test("arrow, page, Home, End and Delete keys", async () => {
	render(<Harness />);
	slider().focus();
	const user = userEvent.setup();
	await user.keyboard("{ArrowRight}");
	expect(slider().getAttribute("aria-valuenow")).toBe("0.5");
	await user.keyboard("{ArrowUp}{ArrowUp}");
	expect(slider().getAttribute("aria-valuenow")).toBe("1.5");
	await user.keyboard("{PageUp}");
	expect(slider().getAttribute("aria-valuenow")).toBe("2.5");
	await user.keyboard("{ArrowLeft}{PageDown}");
	expect(slider().getAttribute("aria-valuenow")).toBe("1");
	await user.keyboard("{End}");
	expect(slider().getAttribute("aria-valuenow")).toBe("10");
	await user.keyboard("{ArrowRight}");
	expect(slider().getAttribute("aria-valuenow")).toBe("10");
	await user.keyboard("{Home}");
	expect(slider().getAttribute("aria-valuenow")).toBe("0.5");
	await user.keyboard("{ArrowLeft}");
	expect(slider().getAttribute("aria-valuenow")).toBe("0.5");
	await user.keyboard("{Delete}");
	expect(slider().getAttribute("aria-valuenow")).toBe("0");
});

test("pointer click sets a half-star value from the x position", () => {
	render(<Harness />);
	const el = slider();
	el.getBoundingClientRect = () => ({
		left: 0,
		width: 200,
		top: 0,
		height: 44,
		right: 200,
		bottom: 44,
		x: 0,
		y: 0,
		toJSON() {},
	});
	fireEvent.click(el, { clientX: 85 });
	expect(el.getAttribute("aria-valuenow")).toBe("4.5");
	fireEvent.click(el, { clientX: 200 });
	expect(el.getAttribute("aria-valuenow")).toBe("10");
});

test("plus and minus buttons step by 0.5", async () => {
	render(<Harness initial={5} />);
	await userEvent.click(screen.getByRole("button", { name: "Gameplay +" }));
	expect(slider().getAttribute("aria-valuenow")).toBe("5.5");
	await userEvent.click(screen.getByRole("button", { name: "Gameplay -" }));
	await userEvent.click(screen.getByRole("button", { name: "Gameplay -" }));
	expect(slider().getAttribute("aria-valuenow")).toBe("4.5");
});

test("disabled widget ignores keys and is not focusable", async () => {
	render(<Harness initial={5} disabled />);
	const el = slider();
	expect(el.getAttribute("aria-disabled")).toBe("true");
	fireEvent.keyDown(el, { key: "ArrowRight" });
	expect(el.getAttribute("aria-valuenow")).toBe("5");
	expect(el.tabIndex).toBe(-1);
});

test("renders no inline style attribute (CSP style-src-attr)", () => {
	const { container } = render(<StarRating label="Gameplay" value={2.5} onChange={() => {}} />);
	expect(container.querySelector("[style]")).toBeNull();
});

test("draws ten SVG stars with full, half and empty fills", () => {
	const { container } = render(<StarRating label="Gameplay" value={2.5} onChange={() => {}} />);
	const fills = [...container.querySelectorAll("[data-fill]")].map((el) =>
		el.getAttribute("data-fill"),
	);
	expect(fills).toEqual([
		"full",
		"full",
		"half",
		"empty",
		"empty",
		"empty",
		"empty",
		"empty",
		"empty",
		"empty",
	]);
	expect(slider().textContent).toBe("");
});
