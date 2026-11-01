import { act, cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";

import Carousel, { SLIDE_MS } from "./Carousel";

vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, string | number>) =>
		p ? `${key} ${JSON.stringify(p)}` : key,
}));
vi.mock("../Nav", () => ({ iconButton: "icon-button" }));

let reduced = false;

beforeEach(() => {
	vi.useFakeTimers();
	reduced = false;
	vi.stubGlobal("matchMedia", () => ({
		matches: reduced,
		addEventListener: () => {},
		removeEventListener: () => {},
	}));
});
afterEach(() => {
	cleanup();
	vi.useRealTimers();
});

const slides = ["a", "b", "c", "d"].map((k) => ({ key: k, node: <p>slide {k}</p> }));
const current = () =>
	screen
		.getAllByRole("group", { hidden: true })
		.findIndex((g) => g.getAttribute("aria-hidden") === "false");
const tick = () => act(() => vi.advanceTimersByTime(SLIDE_MS));

test("auto-advances every 4 s and wraps; inactive slides are hidden and inert", () => {
	render(<Carousel label="Trending games" slides={slides} />);
	const region = screen.getByRole("region", { name: "Trending games" });
	expect(region.getAttribute("aria-roledescription")).toBe("carousel");
	expect(current()).toBe(0);
	const groups = screen.getAllByRole("group", { hidden: true });
	expect(groups[1].hasAttribute("inert")).toBe(true);
	expect(groups[0].hasAttribute("inert")).toBe(false);
	tick();
	expect(current()).toBe(1);
	tick();
	tick();
	tick();
	expect(current()).toBe(0);
});

test("the pause button stops autoplay and play resumes it", () => {
	render(<Carousel label="x" slides={slides} />);
	fireEvent.click(screen.getByRole("button", { name: "home.carousel.pause" }));
	tick();
	tick();
	expect(current()).toBe(0);
	fireEvent.click(screen.getByRole("button", { name: "home.carousel.play" }));
	tick();
	expect(current()).toBe(1);
});

test("hover and keyboard focus pause autoplay", () => {
	render(<Carousel label="x" slides={slides} />);
	const region = screen.getByRole("region");
	fireEvent.mouseEnter(region);
	tick();
	expect(current()).toBe(0);
	fireEvent.mouseLeave(region);
	act(() => screen.getByRole("button", { name: "home.carousel.next" }).focus());
	expect(current()).toBe(0);
	tick();
	expect(current()).toBe(0);
});

test("prefers-reduced-motion: no autoplay until the visitor presses play", () => {
	reduced = true;
	render(<Carousel label="x" slides={slides} />);
	tick();
	expect(current()).toBe(0);
	expect(screen.getByRole("button", { name: "home.carousel.play" })).toBeTruthy();
});

test("arrow keys, previous/next and labelled indicators move between slides", () => {
	reduced = true;
	render(<Carousel label="x" slides={slides} />);
	const region = screen.getByRole("region");
	fireEvent.keyDown(region, { key: "ArrowRight" });
	expect(current()).toBe(1);
	fireEvent.keyDown(region, { key: "ArrowLeft" });
	fireEvent.keyDown(region, { key: "ArrowLeft" });
	expect(current()).toBe(3);
	fireEvent.click(screen.getByRole("button", { name: "home.carousel.next" }));
	expect(current()).toBe(0);
	fireEvent.click(screen.getByRole("button", { name: 'home.carousel.goTo {"n":3}' }));
	expect(current()).toBe(2);
	expect(
		screen
			.getByRole("button", { name: 'home.carousel.goTo {"n":3}' })
			.getAttribute("aria-current"),
	).toBe("true");
	expect(screen.getAllByRole("group", { hidden: true })[2].getAttribute("aria-label")).toBe(
		'home.carousel.slide {"n":3,"total":4}',
	);
});

test("a horizontal swipe changes the slide", () => {
	reduced = true;
	render(<Carousel label="x" slides={slides} />);
	const region = screen.getByRole("region");
	fireEvent.touchStart(region, { touches: [{ clientX: 300 }] });
	fireEvent.touchEnd(region, { changedTouches: [{ clientX: 200 }] });
	expect(current()).toBe(1);
	fireEvent.touchStart(region, { touches: [{ clientX: 100 }] });
	fireEvent.touchEnd(region, { changedTouches: [{ clientX: 120 }] });
	expect(current()).toBe(1);
});
