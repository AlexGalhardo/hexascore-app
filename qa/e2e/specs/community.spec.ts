import { createVerifiedAccount } from "../helpers/account";
import { signIn } from "../helpers/auth";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// P4.B.5.1 quality gate. Needs community_submissions on (E2E_COMMUNITY_OPEN=true with
// FLAG_COMMUNITY_SUBMISSIONS=true on the app), the DemoCatalog admin, and docker access for the signup mail.
// The member side runs in each locale project (axe in en, pt-BR, es); the moderation console is server-rendered
// English HTML, so the admin steps use English labels.
const flagOn = process.env.E2E_COMMUNITY_OPEN === "true";

test.describe("community quality gate", () => {
	test.skip(
		!flagOn || !canReadMail(),
		"needs the community flag and docker access to the app container logs",
	);

	test("suggest, submit, approve, then revert from the audit log", async ({
		page,
		browser,
		prefix,
		t,
		locale,
	}) => {
		test.setTimeout(360_000);
		const title = `E2E gate film ${locale} ${Date.now()}`;

		await createVerifiedAccount(page, prefix, t);
		await page.goto(`${prefix}/suggest/new/movie`);
		await expectNoSeriousA11yViolations(page);
		await page.getByLabel(`${t("community.fields.title")} *`).fill(title);
		await page.getByLabel(t("community.sources")).fill("https://www.wikidata.org/wiki/Q127367");
		await page.getByRole("button", { name: t("community.saveDraft") }).click();
		await expect(page.getByText(t("community.draftRestored"))).toBeVisible();

		await page.goto(`${prefix}/me/suggestions`);
		await page.getByRole("link", { name: title }).click();
		await expectNoSeriousA11yViolations(page);
		await page.getByRole("button", { name: t("community.status.submit") }).click();
		await expect(page.getByTestId("status")).toContainText(
			t("community.status.states.submitted"),
		);

		// A separate context so the member session stays signed in.
		const adminContext = await browser.newContext({ locale: "en" });
		const admin = await adminContext.newPage();
		try {
			await signIn(admin, "");
			await admin.goto("/admin/moderation");
			const row = admin
				.locator("tr", { has: admin.getByRole("link", { name: /movie/ }) })
				.first();
			await row.getByRole("button", { name: "Claim" }).click();
			await admin.getByRole("link", { name: /movie/ }).first().click();
			await admin.getByRole("button", { name: "Approve", exact: true }).click();

			await admin.goto("/admin/catalog-audit");
			await expect(admin.getByRole("link", { name: "suggestion" }).first()).toBeVisible();
			await admin.locator("select[name=reason]").first().selectOption({ index: 1 });
			await admin.getByRole("button", { name: "Revert" }).first().click();
			await expect(admin.getByText("reverted").first()).toBeVisible();
		} finally {
			await adminContext.close();
		}

		await page.goto(`${prefix}/me/suggestions`);
		await expect(page.getByText(title)).toBeVisible();
	});

	// Both moderators must be distinct users; the DemoCatalog seeds only one admin.
	test.fixme("two admins claiming one suggestion: the second claim is refused", () => {});
});
