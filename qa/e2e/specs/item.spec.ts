import { openTopItem } from "../helpers/catalog";
import { expect, expectAccessible, LOCALES, test } from "../helpers/fixtures";

test.describe("item page", () => {
	test("shows score, categories and a centered, paginated users' scores modal", async ({ page }) => {
		await openTopItem(page, "games");
		await expect(page.locator("h1")).toBeVisible();
		await expect(page.getByText("Community score").first()).toBeVisible();
		await expectAccessible(page);

		await page.getByRole("button", { name: "See users' scores" }).click();
		const dialog = page.getByRole("dialog", { name: "Users' scores" });
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole("table")).toBeVisible();
		await expectAccessible(page);

		// centered in the viewport and stacked above the sticky navbar
		const box = await dialog.boundingBox();
		const viewport = page.viewportSize();
		if (!box || !viewport) throw new Error("dialog or viewport has no size");
		expect(Math.abs(box.x + box.width / 2 - viewport.width / 2)).toBeLessThan(2);
		expect(Math.abs(box.y + box.height / 2 - viewport.height / 2)).toBeLessThan(2);
		const onTop = await page.evaluate(
			({ x, y }) => Boolean(document.elementFromPoint(x, y)?.closest('[role="dialog"]')),
			{ x: viewport.width / 2, y: box.y + 4 },
		);
		expect(onTop).toBe(true);

		const pages = dialog.getByRole("navigation", { name: "Ratings pages" });
		await expect(pages).toBeVisible(); // the seed gives the top item ~26 public ratings → 3 pages
		const status = pages.getByText(/^Page \d+ of \d+$/);
		await expect(status).toHaveText(/^Page 1 of [2-9]\d*$/);
		await expect(pages.getByRole("button", { name: "Previous" })).toBeDisabled();
		const firstRow = dialog.locator("tbody tr").first();
		const before = await firstRow.textContent();
		await pages.getByRole("button", { name: "Next" }).click();
		await expect(status).toHaveText(/^Page 2 of \d+$/);
		await expect(firstRow).not.toHaveText(before ?? "");
		await expect(dialog.locator("tbody tr")).not.toHaveCount(0);
		const numbered = pages.getByRole("button", { name: "Go to page 2" });
		if (await numbered.isVisible()) await expect(numbered).toHaveAttribute("aria-current", "page");
		await pages.getByRole("button", { name: "Previous" }).click();
		await expect(status).toHaveText(/^Page 1 of \d+$/);
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
