import { createVerifiedAccount } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// The flow needs community_submissions on (FLAG_COMMUNITY_SUBMISSIONS=true on the app container); without it the
// suggest page shows the closed notice, which the first test covers.
const flagOn = process.env.E2E_COMMUNITY_OPEN === "true";

test.describe("community draft", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("item page links to the suggest page, which is closed or open by flag", async ({
		page,
		prefix,
		t,
	}) => {
		test.setTimeout(240_000);
		await createVerifiedAccount(page, prefix, t);
		await openTopItem(page, prefix, "games", 0);
		await page.getByRole("link", { name: t("community.suggest.link") }).click();
		await expect(page).toHaveURL(/\/suggest\/item\//);
		if (!flagOn) {
			await expect(page.getByText(t("community.closed.title"))).toBeVisible();
			await expectNoSeriousA11yViolations(page);
			return;
		}

		await expect(page.getByTestId("quota-left")).toBeVisible();
		await expectNoSeriousA11yViolations(page);
		const title = page.getByLabel(t("community.fields.title"));
		await title.fill("E2E suggested title");
		await page.getByLabel(t("community.sources")).fill("https://www.wikidata.org/wiki/Q127367");
		await page.getByRole("button", { name: t("community.saveDraft") }).click();
		await expect(page.getByText(t("community.draftRestored"))).toBeVisible();
		await page.reload();
		await expect(page.getByLabel(t("community.fields.title"))).toHaveValue(
			"E2E suggested title",
		);
	});
});
