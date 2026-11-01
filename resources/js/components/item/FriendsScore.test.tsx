import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import FriendsScore from "./FriendsScore";

vi.mock("@inertiajs/react", () => ({ usePage: () => ({ props: { locale: "en" } }) }));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, p?: Record<string, number>) => (p ? `${key}:${p.count}` : key),
}));

afterEach(cleanup);

test("shows the average and who rated", () => {
	render(
		<FriendsScore
			data={{
				count: 2,
				average: 7.5,
				friends: [
					{ nickname: "Ana", score: 8 },
					{ nickname: "Bo", score: 7 },
				],
			}}
		/>,
	);
	const box = screen.getByTestId("friends-score");
	expect(box.textContent).toContain("friendsScore.summary:2");
	expect(box.textContent).toContain("Ana 8.0, Bo 7.0");
});

test("renders nothing without friends", () => {
	render(<FriendsScore data={null} />);
	expect(screen.queryByTestId("friends-score")).toBeNull();
});
