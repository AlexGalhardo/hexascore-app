import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import StatusPicker from "./StatusPicker";

vi.mock("@inertiajs/react", () => ({
	router: { put: vi.fn(), delete: vi.fn() },
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (path: string) => path }));
vi.mock("../../lib/useT", () => ({ useT: () => (key: string) => key }));

afterEach(cleanup);

test("unreleased item keeps only the want status enabled", () => {
	render(<StatusPicker itemId="1" type="game" status={null} unreleased />);
	expect((screen.getByText("social.status.game.want") as HTMLButtonElement).disabled).toBe(false);
	expect((screen.getByText("social.status.game.in_progress") as HTMLButtonElement).disabled).toBe(
		true,
	);
	expect((screen.getByText("social.status.game.completed") as HTMLButtonElement).disabled).toBe(
		true,
	);
	expect((screen.getByText("social.status.game.dropped") as HTMLButtonElement).disabled).toBe(
		true,
	);
});

test("unreleased item still lets the owner clear an existing status", () => {
	render(<StatusPicker itemId="1" type="game" status="in_progress" unreleased />);
	expect((screen.getByText("social.status.game.in_progress") as HTMLButtonElement).disabled).toBe(
		false,
	);
	expect((screen.getByText("social.status.game.completed") as HTMLButtonElement).disabled).toBe(
		true,
	);
});

test("released item keeps every status enabled", () => {
	render(<StatusPicker itemId="1" type="game" status={null} />);
	expect((screen.getByText("social.status.game.completed") as HTMLButtonElement).disabled).toBe(
		false,
	);
});
