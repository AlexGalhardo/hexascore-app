import { act, cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import Carousel from "./Carousel";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));
vi.mock("../Nav", () => ({ iconButton: "icon-button" }));
afterEach(cleanup);

const slides = ["a", "b", "c", "d"].map((key) => ({ key, node: <p>slide {key}</p> }));
const current = () =>
	screen
		.getAllByRole("group", { hidden: true })
		.findIndex((group) => group.getAttribute("aria-hidden") === "false");

test("slides advance on their own every two seconds and pause while the carousel is hovered", () => {
	vi.useFakeTimers();
	try {
		render(<Carousel label="Upcoming games" slides={slides} />);
		expect(current()).toBe(0);
		act(() => vi.advanceTimersByTime(2000));
		expect(current()).toBe(1);
		const region = screen.getByRole("region", { name: "Upcoming games" });
		fireEvent.mouseEnter(region);
		act(() => vi.advanceTimersByTime(4000));
		expect(current()).toBe(1);
		fireEvent.mouseLeave(region);
		act(() => vi.advanceTimersByTime(2000));
		expect(current()).toBe(2);
	} finally {
		vi.useRealTimers();
	}
});

test("arrow keys and swipes navigate the slides, with no buttons", () => {
	render(<Carousel label="Upcoming games" slides={slides} />);
	const region = screen.getByRole("region", { name: "Upcoming games" });
	fireEvent.keyDown(region, { key: "ArrowRight" });
	expect(current()).toBe(1);
	fireEvent.keyDown(region, { key: "ArrowLeft" });
	expect(current()).toBe(0);
	fireEvent.touchStart(region, { touches: [{ clientX: 300 }] });
	fireEvent.touchEnd(region, { changedTouches: [{ clientX: 200 }] });
	expect(current()).toBe(1);
	expect(screen.queryByRole("button")).toBeNull();
});
