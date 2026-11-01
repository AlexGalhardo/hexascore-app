import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem, SECTIONS, type Section } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { escapeRegExp } from "../helpers/t";

// Critic scores are dropped (D-09).
test.describe("item page", () => {
	for (const section of SECTIONS) {
		test(`${section}: shows the community score with category averages and attribution`, async ({
			page,
			prefix,
			locale,
			t,
		}) => {
			await openTopItem(page, prefix, section);
			await expect(page.locator("html")).toHaveAttribute(
				"lang",
				new RegExp(`^${locale.split("-")[0]}`, "i"),
			);
			await expect(page.locator("h1")).toBeVisible();
			await expect(page.getByTestId("score-block")).toBeVisible();
			await expect(
				page.getByRole("heading", { name: t("item.categoryScores") }),
			).toBeAttached();
			await expect(page.getByText(t("item.attribution"))).toBeVisible();
			await expectNoSeriousA11yViolations(page);
		});
	}

	test("header shows the cover and score badge; the rating panel is above the fold on mobile", async ({
		page,
		prefix,
	}, testInfo) => {
		await openTopItem(page, prefix, "games");
		const header = page.locator("article > header");
		await expect(header.getByTestId("score-block")).toBeVisible();
		await expect(header.locator("img, [data-testid='cover-placeholder']")).toHaveCount(1);

		if (testInfo.project.name.startsWith("mobile")) {
			const heading = page.locator("#rating-heading");
			const box = await heading.boundingBox();
			expect(box?.y ?? Infinity).toBeLessThan(page.viewportSize()?.height ?? 0);
		}
	});

	test("a guest is asked to log in to rate", async ({ page, prefix, t }) => {
		await openTopItem(page, prefix, "games");
		await expect(page.getByRole("link", { name: t("item.loginToRate") })).toBeVisible();
		await expect(page.locator("input[type=radio]:enabled")).toHaveCount(0);
	});

	test("an unknown slug is a 301 to the home page", async ({ request, prefix }) => {
		const response = await request.get(`${prefix}/games/does-not-exist-0000`, {
			maxRedirects: 0,
		});
		expect(response.status()).toBe(301);
		expect(response.headers().location).toBe(`${prefix}/`);
	});

	test("shows score, categories and a centered, paginated users' scores modal", async ({
		page,
		prefix,
		t,
	}) => {
		await openTopItem(page, prefix, "games");
		await expect(page.getByRole("heading", { name: t("item.communityScore") })).toBeVisible();
		await expectNoSeriousA11yViolations(page);

		await page.getByRole("button", { name: t("item.seeUserScores") }).click();
		const dialog = page.getByRole("dialog", { name: t("item.userScores") });
		await expect(dialog).toBeVisible();
		await expect(dialog.getByRole("table")).toBeVisible();
		await expectNoSeriousA11yViolations(page);

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

		const pages = dialog.getByRole("navigation", { name: t("item.ratingsPages") });
		await expect(pages).toBeVisible();
		const status = pages.getByText(pageLine(t, 1, "[2-9]\\d*"));
		await expect(status).toBeVisible();
		await expect(pages.getByRole("button", { name: t("item.previous") })).toBeDisabled();
		const firstRow = dialog.locator("tbody tr").first();
		const before = await firstRow.textContent();
		await pages.getByRole("button", { name: t("item.next") }).click();
		await expect(pages.getByText(pageLine(t, 2, "\\d+"))).toBeVisible();
		await expect(firstRow).not.toHaveText(before ?? "");
		await expect(dialog.locator("tbody tr")).not.toHaveCount(0);
		const numbered = pages.getByRole("button", { name: t("item.goToPage", { page: 2 }) });
		if (await numbered.isVisible())
			await expect(numbered).toHaveAttribute("aria-current", "page");
		await pages.getByRole("button", { name: t("item.previous") }).click();
		await expect(pages.getByText(pageLine(t, 1, "\\d+"))).toBeVisible();
		await expectNoSeriousA11yViolations(page);

		await page.keyboard.press("Escape");
		await expect(dialog).toBeHidden();
	});

	test("trailers load only after consent, from youtube-nocookie", async ({ page, prefix, t }) => {
		const youtube: string[] = [];
		await page.route(/youtube/, (route) => {
			youtube.push(route.request().url());
			return route.fulfill({ status: 204, body: "" });
		});
		for (const section of ["movies", "games", "series"] as const satisfies readonly Section[]) {
			await openTopItem(page, prefix, section);
			if (await page.getByRole("heading", { name: t("item.trailers") }).count()) break;
		}
		const load = page
			.getByRole("button", { name: new RegExp(escapeRegExp(t("item.loadTrailer"))) })
			.first();
		test.skip(!(await load.count()), "no seeded item with trailers among the top items");
		expect(youtube).toEqual([]);
		await load.click();
		const frame = page
			.locator('iframe[src^="https://www.youtube-nocookie.com/embed/"]')
			.first();
		await expect(frame).toBeVisible();
		expect(youtube.every((url) => url.startsWith("https://www.youtube-nocookie.com/"))).toBe(
			true,
		);
	});

	test.describe("with embeds allowed", () => {
		test.use({ consent: "all" });

		test("trailers load immediately", async ({ page, prefix, t }) => {
			await page.route(/youtube/, (route) => route.fulfill({ status: 204, body: "" }));
			await openTopItem(page, prefix, "movies");
			const heading = page.getByRole("heading", { name: t("item.trailers") });
			test.skip(!(await heading.count()), "top movie has no seeded trailers");
			await expect(
				page.locator('iframe[src^="https://www.youtube-nocookie.com/embed/"]').first(),
			).toBeVisible();
		});
	});
});

function pageLine(
	t: (key: string, params?: Record<string, string | number>) => string,
	page: number,
	count: string,
): RegExp {
	const sample = t("item.pageOf", { page, count: "COUNT" });
	return new RegExp(escapeRegExp(sample).replace("COUNT", count));
}
