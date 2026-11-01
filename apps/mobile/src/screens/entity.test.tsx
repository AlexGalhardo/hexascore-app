import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react-native";
import type { AppItemJSON } from "../../../../packages/app-client/src/api";
import { type EntityReader, EntityView } from "./Entity";

const item = (type: string, n: number): AppItemJSON => ({
	id: `${type}-${n}`,
	slug: `${type}-slug-${n}`,
	type,
	locale: "en",
	title: `${type} title ${n}`,
	score: 7.5,
	ratingCount: 9,
	releaseDate: "2001-05-01",
});

function reader(byType: Record<string, AppItemJSON[]>): EntityReader & {
	listEntityItems: jest.Mock;
} {
	return {
		getEntity: jest.fn(async () => ({
			id: "e1",
			slug: "frank-herbert",
			kind: "person",
			locale: "en",
			name: "Frank Herbert",
			bornOn: "1920-10-08",
		})),
		listEntityItems: jest.fn(async (args: { query?: { type?: string } }) => {
			const data = byType[args.query?.type ?? ""] ?? [];
			return { data, meta: { total: data.length } };
		}),
	};
}

async function renderEntity(read: EntityReader, onOpen = jest.fn()) {
	const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	await render(
		<QueryClientProvider client={queryClient}>
			<EntityView id="e1" read={read} locale="en" onOpen={onOpen} onBrowse={() => {}} />
		</QueryClientProvider>,
	);
	return onOpen;
}

describe("entity screen", () => {
	test("groups follow ITEM_TYPES order and empty groups are hidden", async () => {
		const read = reader({ books: [item("books", 1)], games: [item("games", 1), item("games", 2)] });
		await renderEntity(read);
		expect(await screen.findByRole("header", { name: "Frank Herbert" })).toBeTruthy();
		await screen.findByText("books title 1");
		const headers = screen.getAllByRole("header").map((h) => h.props.children);
		expect(headers).toEqual(["Frank Herbert", "Games", "Books"]);
		expect(screen.queryByRole("header", { name: "Movies" })).toBeNull();
		for (const type of ["games", "movies", "series", "books"]) {
			expect(read.listEntityItems).toHaveBeenCalledWith({
				params: { id: "e1" },
				query: { type, locale: "en" },
			});
		}
	});

	test("an entity with no items shows the empty state", async () => {
		await renderEntity(reader({}));
		expect(await screen.findByText("No public items yet.")).toBeTruthy();
	});

	test("tapping a row opens the item route", async () => {
		const onOpen = await renderEntity(reader({ games: [item("games", 1)] }));
		await fireEvent.press(await screen.findByRole("button", { name: /games title 1/ }));
		expect(onOpen).toHaveBeenCalledWith("/games/games-slug-1?id=games-1");
	});
});
