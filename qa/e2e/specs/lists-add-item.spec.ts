import { createVerifiedAccount, open } from "../helpers/account";
import { openTopItem } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

test.describe("list edit page", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("the owner adds an item from the edit page and sees it in the public list", async ({
		page,
		prefix,
		t,
	}, testInfo) => {
		test.setTimeout(240_000);
		const account = await createVerifiedAccount(page, prefix, t);
		const title = `E2E add ${testInfo.project.name}`;

		await openTopItem(page, prefix, "games", 0);
		const query = ((await page.locator("h1").textContent()) ?? "").trim().slice(0, 4);

		await open(page, `${prefix}/lists`);
		await page.getByLabel(t("userLists.titleField")).fill(title);
		await page.getByRole("button", { name: t("userLists.create") }).click();
		const editHref = (await page.getByRole("link", { name: title }).getAttribute("href")) ?? "";
		const id = editHref.split("/").at(-2) ?? "";

		await open(page, editHref);
		await page.getByRole("combobox", { name: t("userLists.addItem") }).fill(query);
		const option = page.getByRole("option").first();
		await expect(option).toBeVisible();
		const name = ((await option.locator("span").first().textContent()) ?? "").trim();
		await option.click();
		await expect(page.locator("main ol").getByRole("link", { name })).toBeVisible();

		await open(page, `${prefix}/u/${account.nickname}/lists/${id}`);
		await expect(page.getByRole("link", { name })).toBeVisible();
	});
});
