import { createVerifiedAccount, open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

test.describe("lists", () => {
	test("lists require login", async ({ page, prefix }) => {
		await open(page, `${prefix}/lists`);
		await expect(page).toHaveURL(/\/login/);
	});

	test.describe("with a verified account", () => {
		test.skip(
			!canReadMail(),
			"needs docker access to the app container logs (E2E_APP_CONTAINER)",
		);

		test("create, fill, reorder, publish, hide and delete a list", async ({
			page,
			browser,
			prefix,
			t,
		}, testInfo) => {
			test.setTimeout(240_000);
			const account = await createVerifiedAccount(page, prefix, t);
			const title = `E2E list ${testInfo.project.name}`;

			await open(page, `${prefix}/lists`);
			await expect(page.getByText(t("lists.noLists"))).toBeVisible();
			await expectNoSeriousA11yViolations(page);
			await page.getByLabel(t("userLists.titleField")).fill(title);
			await page.getByRole("button", { name: t("userLists.create") }).click();
			const link = page.getByRole("link", { name: title });
			await expect(link).toBeVisible();
			const editHref = (await link.getAttribute("href")) ?? "";
			const id = editHref.split("/").at(-2) ?? "";
			expect(editHref).toContain(`${prefix}/lists/`);

			// Add two games from their pages.
			const added: string[] = [];
			for (const nth of [0, 1]) {
				await openTopItem(page, prefix, "games", nth);
				added.push(((await page.locator("h1").textContent()) ?? "").trim());
				const row = page.locator("li", { hasText: title }).filter({
					has: page.getByRole("button", { name: t("lists.addToList") }),
				});
				await row.getByRole("button", { name: t("lists.addToList") }).click();
				await expect(
					page
						.locator("li", { hasText: title })
						.getByRole("button", { name: t("lists.remove") }),
				).toBeVisible();
			}

			// Reorder: move the second item up and save.
			await open(page, editHref);
			await expectNoSeriousA11yViolations(page);
			await page
				.getByRole("button", { name: `${t("userLists.moveUp")}: ${added[1]}` })
				.click();
			// Reload only after the save landed, or the page can show the old order.
			const saved = page.waitForResponse(
				(r) => r.url().endsWith("/order") && r.request().method() === "PUT",
			);
			await page.getByRole("button", { name: t("userLists.saveOrder") }).click();
			await saved;
			await page.reload();
			const names = await page.locator("main li a[href*='/games/']").allInnerTexts();
			const first = names.findIndex((n) => n.includes(added[1]));
			const second = names.findIndex((n) => n.includes(added[0]));
			expect(first).toBeGreaterThanOrEqual(0);
			expect(first).toBeLessThan(second);

			// A public list is visible to anyone; a private one is not.
			const publicUrl = `${prefix}/u/${account.nickname}/lists/${id}`;
			const visitor = await (
				await browser.newContext({ locale: testInfo.project.use.locale })
			).newPage();
			expect((await visitor.goto(publicUrl))?.status()).toBe(200);
			await expect(visitor.getByRole("heading", { name: title })).toBeVisible();

			await page.getByLabel(t("userLists.public")).uncheck();
			await page.getByRole("button", { name: t("userLists.save"), exact: true }).click();
			await expect
				.poll(async () => (await visitor.goto(publicUrl))?.status(), { timeout: 15_000 })
				.toBe(404);
			await visitor.context().close();

			// Delete.
			page.once("dialog", (d) => d.accept());
			await page.getByRole("button", { name: t("userLists.delete") }).click();
			await expect(page).toHaveURL(/\/lists$/);
			await expect(page.getByText(t("lists.noLists"))).toBeVisible();
		});
	});
});
