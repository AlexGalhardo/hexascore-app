import { cleanup, fireEvent, render, screen } from "@testing-library/react";
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

test("slides stay still until a visitor navigates, without pause or play buttons", () => {
	vi.useFakeTimers();
	try {
		render(<Carousel label="Upcoming games" slides={slides} />);
		expect(current()).toBe(0);
		vi.advanceTimersByTime(12000);
		expect(current()).toBe(0);
		expect(screen.queryByRole("button", { name: /home.carousel.(pause|play)/ })).toBeNull();
		expect(screen.getAllByRole("group", { hidden: true })[1].hasAttribute("inert")).toBe(true);
	} finally {
		vi.useRealTimers();
	}
});

test("arrow keys, swipes and the revealed arrow buttons navigate the slides", () => {
	render(<Carousel label="Upcoming games" slides={slides} />);
	const region = screen.getByRole("region", { name: "Upcoming games" });
	fireEvent.keyDown(region, { key: "ArrowRight" });
	expect(current()).toBe(1);
	fireEvent.keyDown(region, { key: "ArrowLeft" });
	expect(current()).toBe(0);
	fireEvent.touchStart(region, { touches: [{ clientX: 300 }] });
	fireEvent.touchEnd(region, { changedTouches: [{ clientX: 200 }] });
	expect(current()).toBe(1);
	fireEvent.click(screen.getByRole("button", { name: "home.carousel.next" }));
	expect(current()).toBe(2);
	fireEvent.click(screen.getByRole("button", { name: "home.carousel.previous" }));
	expect(current()).toBe(1);
});
