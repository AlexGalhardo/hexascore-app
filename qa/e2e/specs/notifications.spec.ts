import { createVerifiedAccount, open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";
import { canReadMail, seedNotification } from "../helpers/mail";

test.describe("notifications", () => {
	test.skip(!canReadMail(), "needs docker access to the app and database containers");

	test("an unread notification shows in the inbox, then reads one by one and all at once", async ({
		page,
		prefix,
		t,
	}) => {
		const account = await createVerifiedAccount(page, prefix, t);
		await open(page, `${prefix}/me/notifications`);
		await expect(page.getByText(t("notifications.empty"))).toBeVisible();

		seedNotification(account.email, "system", `e2e-a-${account.nickname}`);
		seedNotification(account.email, "list_hidden", `e2e-b-${account.nickname}`);
		await open(page, `${prefix}/me/notifications`);
		await expectNoSeriousA11yViolations(page);
		await expect(page.getByText(t("notifications.kind.system"))).toBeVisible();
		await expect(page.getByText(t("notifications.kind.list_hidden"))).toBeVisible();

		const markRead = page.getByRole("button", { name: t("notifications.markRead") });
		await expect(markRead).toHaveCount(2);
		await markRead.first().click();
		await expect(markRead).toHaveCount(1);

		await page.getByRole("button", { name: t("notifications.markAllRead") }).click();
		await expect(markRead).toHaveCount(0);
		await expect(
			page.getByRole("button", { name: t("notifications.markAllRead") }),
		).toBeHidden();
	});

	test("a notification of another account is never listed", async ({ page, prefix, t }) => {
		await createVerifiedAccount(page, prefix, t);
		seedNotification("admin@example.test", "system", `e2e-admin-${Date.now()}`);
		await open(page, `${prefix}/me/notifications`);
		await expect(page.getByText(t("notifications.empty"))).toBeVisible();
	});
});
