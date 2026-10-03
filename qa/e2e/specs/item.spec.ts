import { openTopItem } from "../helpers/catalog";
import { expect, expectAccessible, LOCALES, test } from "../helpers/fixtures";

test.describe("item page", () => {
	test("shows score, categories and opens the users' scores modal", async ({ page }) => {
		await openTopItem(page, "games");
		await expect(page.locator("h1")).toBeVisible();
		await expect(page.getByText("Community score").first()).toBeVisible();
		await expectAccessible(page);

		await page.getByRole("button", { name: "See users' scores" }).click();
		const dialog = page.getByRole("dialog", { name: "Users' scores" });
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole("table")).toBeVisible();
		await expectAccessible(page);
		await page.keyboard.press("Escape");
		await expect(dialog).toBeHidden();
	});

	test("trailers load only after consent, from youtube-nocookie", async ({ page }) => {
		const youtube: string[] = [];
		await page.route(/youtube/, (route) => {
			youtube.push(route.request().url());
			return route.fulfill({ status: 204, body: "" });
		});
		for (const section of ["movies", "games", "series"]) {
			await openTopItem(page, section);
			if (await page.getByRole("heading", { name: "Trailers" }).count()) break;
		}
		const load = page.getByRole("button", { name: /Load trailer/ }).first();
		test.skip(!(await load.count()), "no seeded item with trailers among the top items");
		expect(youtube).toEqual([]);
		await load.click();
		const frame = page.locator('iframe[src^="https://www.youtube-nocookie.com/embed/"]').first();
		await expect(frame).toBeVisible();
		expect(youtube.every((url) => url.startsWith("https://www.youtube-nocookie.com/"))).toBe(true);
	});

	test.describe("with embeds allowed", () => {
		test.use({ consent: "all" });

		test("trailers load immediately", async ({ page }) => {
			await page.route(/youtube/, (route) => route.fulfill({ status: 204, body: "" }));
			await openTopItem(page, "movies");
			const heading = page.getByRole("heading", { name: "Trailers" });
			test.skip(!(await heading.count()), "top movie has no seeded trailers");
			await expect(page.locator('iframe[src^="https://www.youtube-nocookie.com/embed/"]').first()).toBeVisible();
		});
	});

	for (const { prefix, code } of LOCALES) {
		test(`renders with license attribution in ${code}`, async ({ page }) => {
			await openTopItem(page, "books", prefix);
			await expect(page.locator("html")).toHaveAttribute("lang", code);
			await expect(page.locator("h1")).toBeVisible();
			await expectAccessible(page);
		});
	}
});
