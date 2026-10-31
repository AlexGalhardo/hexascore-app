import { createVerifiedAccount } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// Needs community_submissions on (E2E_COMMUNITY_OPEN=true with FLAG_COMMUNITY_SUBMISSIONS=true on the app).
const flagOn = process.env.E2E_COMMUNITY_OPEN === "true";

test.describe("community status and my suggestions", () => {
	test.skip(
		!flagOn || !canReadMail(),
		"needs the community flag and docker access to the app container logs",
	);

	test("draft shows diff markers and timeline, then can be withdrawn", async ({
		page,
		prefix,
		t,
	}) => {
		test.setTimeout(240_000);
		await createVerifiedAccount(page, prefix, t);
		await page.goto(`${prefix}/suggest/new/movie`);
		await page.getByLabel(`${t("community.fields.title")} *`).fill("E2E status film");
		await page.getByLabel(t("community.sources")).fill("https://www.wikidata.org/wiki/Q127367");
		await page.getByRole("button", { name: t("community.saveDraft") }).click();
		await expect(page.getByText(t("community.draftRestored"))).toBeVisible();

		await page.goto(`${prefix}/me/suggestions`);
		await page.getByRole("link", { name: "E2E status film" }).click();
		await expect(page.getByText("E2E status film")).toBeVisible();
		await expect(page.getByText(t("community.status.timeline"))).toBeVisible();
		await expectNoSeriousA11yViolations(page);

		await page.getByRole("button", { name: t("community.status.withdraw") }).click();
		await expect(page.getByTestId("status")).toContainText(
			t("community.status.states.withdrawn"),
		);
	});
});
