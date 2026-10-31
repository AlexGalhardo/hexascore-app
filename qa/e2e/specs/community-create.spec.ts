import { createVerifiedAccount } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// Needs community_submissions on (E2E_COMMUNITY_OPEN=true with FLAG_COMMUNITY_SUBMISSIONS=true on the app).
const flagOn = process.env.E2E_COMMUNITY_OPEN === "true";

test.describe("community create and merge request", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("new-record form saves a draft; the duplicate tab renders", async ({
		page,
		prefix,
		t,
	}) => {
		test.setTimeout(240_000);
		await createVerifiedAccount(page, prefix, t);
		await page.goto("/suggest/new/movie");
		if (!flagOn) {
			await expect(page.getByText(t("community.closed.title"))).toBeVisible();
			await expectNoSeriousA11yViolations(page);
			return;
		}

		await expect(page.getByTestId("quota-left")).toBeVisible();
		await expectNoSeriousA11yViolations(page);
		await page.getByLabel(`${t("community.fields.title")} *`).fill("E2E brand new film");
		await page.getByLabel(t("community.sources")).fill("https://www.wikidata.org/wiki/Q127367");
		await page.getByRole("button", { name: t("community.saveDraft") }).click();
		await expect(page.getByText(t("community.draftRestored"))).toBeVisible();

		await page.getByRole("link", { name: t("community.create.reportDuplicate") }).click();
		await expect(page.getByLabel(t("community.create.mergeSource"))).toBeVisible();
		await expectNoSeriousA11yViolations(page);
	});
});
