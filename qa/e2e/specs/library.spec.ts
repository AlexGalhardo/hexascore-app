import { createVerifiedAccount } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem, projectSlot } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

test.describe("library", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("a set status shows the item as a card under its status heading", async ({
		page,
		prefix,
		t,
		any,
	}, testInfo) => {
		await createVerifiedAccount(page, prefix, any);
		await openTopItem(page, prefix, "movies", projectSlot(testInfo));
		const title = ((await page.locator("h1").textContent()) ?? "").trim();

		const completed = page
			.getByRole("group", { name: t("social.status.label") })
			.getByRole("button", { name: t("social.status.movie.completed") });
		await Promise.all([
			page.waitForResponse(
				(r) => r.request().method() === "PUT" && /\/status$/.test(r.url()),
			),
			completed.click(),
		]);
		await expect(completed).toHaveAttribute("aria-pressed", "true");

		await page.goto(`${prefix}/library`);
		const shelf = page.getByRole("region", {
			name: new RegExp(t("social.status.shelves.completed")),
		});
		await expect(shelf.getByRole("link", { name: title })).toBeVisible();
		await expect(shelf.getByText(t("library.notRated"))).toBeVisible();
		await expectNoSeriousA11yViolations(page);
	});
});
