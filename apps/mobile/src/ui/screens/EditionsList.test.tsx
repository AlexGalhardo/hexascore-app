import { render, screen } from "@testing-library/react-native";
import { EditionsList } from "./EditionsList";

describe("EditionsList", () => {
	test("title and one row per edition with its facts", async () => {
		await render(
			<EditionsList
				title="Editions"
				editions={[
					{
						id: "e1",
						title: "First Ace paperback edition",
						facts: [
							{ label: "Published", value: "July 1984" },
							{ label: "ISBN", value: "978-0441569595" },
						],
					},
				]}
			/>,
		);
		expect(screen.getByRole("header", { name: "Editions" })).toBeTruthy();
		expect(screen.getByText("First Ace paperback edition")).toBeTruthy();
		expect(screen.getByText("978-0441569595")).toBeTruthy();
	});

	test("renders nothing without editions", async () => {
		await render(<EditionsList title="Editions" editions={[]} />);
		expect(screen.queryByText("Editions")).toBeNull();
	});
});
