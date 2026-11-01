import { createVerifiedAccount } from "../helpers/account";
import { signIn } from "../helpers/auth";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { DEMO_ADMIN_EMAIL, DEMO_SECOND_ADMIN_EMAIL } from "../helpers/db";
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

	// Needs an open suggestion in the queue (the member test above or the CommunityLoad seeder leaves one).
	test("two admins claiming one suggestion: the second claim is refused", async ({ browser }) => {
		test.setTimeout(120_000);
		const first = await browser.newContext({ locale: "en" });
		const second = await browser.newContext({ locale: "en" });

		try {
			const a = await first.newPage();
			const b = await second.newPage();
			await signIn(a, "", DEMO_ADMIN_EMAIL);
			await signIn(b, "", DEMO_SECOND_ADMIN_EMAIL);
			await a.goto("/admin/moderation");
			await b.goto("/admin/moderation");

			const action = await a.locator("form[action$='/claim']").first().getAttribute("action");
			expect(action).not.toBeNull();
			const claim = (page: typeof a) =>
				page.locator(`form[action='${action}'] button`).click();

			const won = a.waitForResponse((r) => r.url().endsWith(action ?? ""));
			await claim(a);
			expect((await won).status()).toBeLessThan(400);

			const lost = b.waitForResponse((r) => r.url().endsWith(action ?? ""));
			await claim(b);
			const status = (await lost).status();
			expect(status).toBeGreaterThanOrEqual(400);
			expect(status).toBeLessThan(500);
		} finally {
			await first.close();
			await second.close();
		}
	});
});
