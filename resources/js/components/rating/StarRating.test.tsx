import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { afterEach, expect, test, vi } from "vitest";

import StarRating from "./StarRating";

afterEach(cleanup);

vi.mock("@inertiajs/react", () => ({ usePage: () => ({ props: { locale: "en" } }) }));

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, params?: Record<string, string | number>) =>
		({
			"rating.valuetext": `${params?.value} of 10 stars`,
			"rating.outOf": `${params?.value} / 10`,
			"rating.preview": `Preview ${params?.value}`,
			"rating.unset": "Not rated",
		})[key] ?? key,
}));

function Harness({ initial = null }: { initial?: number | null }) {
	const [v, setV] = useState<number | null>(initial);
	return <StarRating label="Gameplay" value={v} onChange={setV} />;
}

const group = () => screen.getByRole("radiogroup", { name: "Gameplay" });
const radio = (v: number) => screen.getByRole("radio", { name: `${v} of 10 stars` });
const printed = () => screen.getByTestId("rating-value").textContent;

test("one radio group of 20 half-star options with the value printed beside it", () => {
	const { rerender } = render(<StarRating label="Gameplay" value={null} onChange={() => {}} />);
	expect(screen.getAllByRole("radio")).toHaveLength(20);
	expect(printed()).toBe("Not rated");
	rerender(<StarRating label="Gameplay" value={7.5} onChange={() => {}} />);
	expect((radio(7.5) as HTMLInputElement).checked).toBe(true);
	expect(printed()).toBe("7.5 / 10");
	expect(group().querySelectorAll('[data-fill="full"]')).toHaveLength(7);
	expect(group().querySelectorAll('[data-fill="half"]')).toHaveLength(1);
});

test("arrows move by 0.5 without wrapping; Home, End and Delete", async () => {
	const user = userEvent.setup();
	render(<Harness initial={9.5} />);
	radio(9.5).focus();
	await user.keyboard("{ArrowRight}");
	expect(printed()).toBe("10.0 / 10");
	expect(document.activeElement).toBe(radio(10));
	await user.keyboard("{ArrowUp}");
	expect(printed()).toBe("10.0 / 10");
	await user.keyboard("{Home}");
	expect(printed()).toBe("0.5 / 10");
	await user.keyboard("{ArrowLeft}");
	expect(printed()).toBe("0.5 / 10");
	await user.keyboard("{End}{ArrowDown}");
	expect(printed()).toBe("9.5 / 10");
	await user.keyboard("{Delete}");
	expect(printed()).toBe("Not rated");
});

test("clicking a half star selects it; hover previews", async () => {
	const user = userEvent.setup();
	render(<Harness />);
	await user.click(radio(3.5));
	expect(printed()).toBe("3.5 / 10");
	fireEvent.pointerEnter(radio(6.5).closest("label") as HTMLElement);
	expect(printed()).toBe("Preview 6.5");
	fireEvent.pointerLeave(group());
	expect(printed()).toBe("3.5 / 10");
});

test("disabled stars do not change", () => {
	const onChange = vi.fn();
	render(<StarRating label="Gameplay" value={null} onChange={onChange} disabled />);
	fireEvent.keyDown(group(), { key: "End" });
	expect(onChange).not.toHaveBeenCalled();
	expect((radio(5) as HTMLInputElement).disabled).toBe(true);
});
