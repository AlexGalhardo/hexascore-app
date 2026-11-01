import { createVerifiedAccount, signOut } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { openTopItem, projectSlot } from "../helpers/catalog";
import { expect, test } from "../helpers/fixtures";
import { canReadMail } from "../helpers/mail";

// Report-a-user has no UI yet (see admin.spec.ts). Follows need the social layer on: E2E_SOCIAL_OPEN=true with
// FLAG_SOCIAL=true on the app (docker-compose.yml passes it through); both default off (SDD 14 §12.3, Q-A5).
const socialOn = process.env.E2E_SOCIAL_OPEN === "true";
test.describe("profiles", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");

	test("a public profile lists ratings; a private one hides them from everyone but the owner", async ({
		page,
		prefix,
		t,
		any,
	}, testInfo) => {
		const account = await createVerifiedAccount(page, prefix, any);
		// Slot offset keeps this spec off the items the ratings spec rates in the same run.
		const itemHref = await openTopItem(page, prefix, "books", projectSlot(testInfo));
		const title = ((await page.locator("h1").textContent()) ?? "").trim();

		const groups = page.getByRole("radiogroup");
		for (let i = 0; i < 6; i++) {
			await groups.nth(i).getByRole("radio").first().focus();
			await page.keyboard.press("End");
		}
		await page.getByRole("button", { name: t("ratings.save") }).click();
		await expect(page.getByRole("button", { name: any("ratings.update") })).toBeVisible();

		const profile = `${prefix}/u/${account.nickname.toLowerCase()}`;
		await page.goto(profile);
		await expect(page.getByRole("heading", { level: 1, name: account.nickname })).toBeVisible();
		await expect(page.getByRole("link", { name: title }).first()).toBeVisible();
		await expectNoSeriousA11yViolations(page);

		await page.goto(`${prefix}/profile`);
		const privacy = page.locator("#privacy");
		await privacy.getByLabel(t("account.privacy.label")).check();
		const saved = page.waitForResponse(
			(r) => r.url().endsWith("/profile/privacy") && r.request().method() === "PATCH",
		);
		await privacy.locator("button[type=submit]").click();
		expect((await saved).status()).toBeLessThan(400);

		await page.goto(profile);
		await expect(page.getByText(t("social.profile.privateOwner"))).toBeVisible();
		await expect(page.getByRole("link", { name: title }).first()).toBeVisible();

		await signOut(page);
		await page.goto(profile);
		await expect(page.getByText(t("social.profile.private"))).toBeVisible();
		await expect(page.getByRole("link", { name: title })).toHaveCount(0);
		await expectNoSeriousA11yViolations(page);
		expect(itemHref).toContain(`${prefix}/books/`);
	});
});

test.describe("follows", () => {
	test.skip(!canReadMail(), "needs docker access to the app container logs (E2E_APP_CONTAINER)");
	test.skip(!socialOn, "needs FLAG_SOCIAL=true on the app and E2E_SOCIAL_OPEN=true");

	test("follow and unfollow update the counts; a private profile turns a follow into a request the owner accepts", async ({
		page,
		browser,
		prefix,
		t,
		any,
	}, testInfo) => {
		test.setTimeout(300_000);
		const star = await createVerifiedAccount(page, prefix, any);
		// A second member in its own context and client IP, so the per-IP signup and login limits stay apart.
		const fanContext = await browser.newContext({
			locale: testInfo.project.use.locale,
			extraHTTPHeaders: {
				"X-Forwarded-For": `10.250.${1 + testInfo.workerIndex}.${1 + testInfo.repeatEachIndex}`,
			},
		});
		const fan = await fanContext.newPage();
		const fanAccount = await createVerifiedAccount(fan, prefix, any);
		const starProfile = `${prefix}/u/${star.nickname.toLowerCase()}`;
		const followers = (count: number): string => t("social.follow.followers", { count });

		await fan.goto(starProfile);
		await expect(fan.getByText(followers(0))).toBeVisible();
		await fan.getByRole("button", { name: t("social.follow.follow") }).click();
		await expect(fan.getByText(followers(1))).toBeVisible();
		await expect(fan.getByRole("button", { name: t("social.follow.unfollow") })).toBeVisible();
		await expectNoSeriousA11yViolations(fan);

		await fan.getByRole("button", { name: t("social.follow.unfollow") }).click();
		await expect(fan.getByText(followers(0))).toBeVisible();

		await page.goto(`${prefix}/profile`);
		const privacy = page.locator("#privacy");
		await privacy.getByLabel(t("account.privacy.label")).check();
		const saved = page.waitForResponse(
			(r) => r.url().endsWith("/profile/privacy") && r.request().method() === "PATCH",
		);
		await privacy.locator("button[type=submit]").click();
		expect((await saved).status()).toBeLessThan(400);

		await fan.goto(starProfile);
		await expect(fan.getByText(followers(0))).toHaveCount(0);
		await fan.getByRole("button", { name: t("social.follow.follow") }).click();
		await expect(fan.getByRole("button", { name: t("social.follow.requested") })).toBeVisible();

		await page.goto(`${prefix}/me/notifications`);
		await expect(
			page.getByText(
				t("notifications.kind.follow_request", { nickname: fanAccount.nickname }),
			),
		).toBeVisible();
		await expectNoSeriousA11yViolations(page);
		await page.getByRole("button", { name: t("notifications.followAccept") }).click();

		await page.goto(starProfile);
		await expect(page.getByText(followers(1))).toBeVisible();
		await fan.goto(starProfile);
		await expect(fan.getByRole("button", { name: t("social.follow.unfollow") })).toBeVisible();
		await fanContext.close();
	});
});
