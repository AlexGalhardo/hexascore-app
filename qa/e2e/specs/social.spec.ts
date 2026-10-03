import { createVerifiedAccount, expectLoggedIn, hasAuthSecret, logIn } from "../helpers/auth";
import { createItem, createMember, findUser, hasDatabase, insertRating } from "../helpers/db";
import { expect, expectAccessible, test } from "../helpers/fixtures";
import { logOut } from "../helpers/session";

test.describe("profiles and feed @auth", () => {
	test.skip(!hasAuthSecret() || !hasDatabase(), "needs E2E_BETTER_AUTH_SECRET and E2E_DATABASE_URL");

	test("public profile lists ratings; a private one hides them from everyone but the owner", async ({
		page,
		request,
		baseURL,
	}) => {
		const item = await createItem();
		const account = await createVerifiedAccount(request, baseURL ?? "", "privacy");
		const id = (await findUser(account.email))?.id;
		if (!id) throw new Error("account not found");
		await insertRating(id, item.id, [70, 72, 74, 76, 78, 80]);
		const profile = `/u/${account.nickname}`;

		await page.goto(profile);
		await expect(page.getByRole("heading", { level: 1, name: account.nickname })).toBeVisible();
		await expect(page.getByRole("link", { name: item.title })).toBeVisible();
		await expect(page.getByRole("img", { name: "Score 75 out of 100" }).first()).toBeVisible();
		await expectAccessible(page);

		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto("/profile");
		const toggle = page.getByRole("switch");
		await toggle.check();
		await expect(toggle).toBeEnabled();
		await expect(toggle).toBeChecked();

		await page.goto(profile);
		await expect(page.getByText(/Private profile \(only you see this page\)/)).toBeVisible();
		await expect(page.getByRole("link", { name: item.title })).toBeVisible();

		await logOut(page);
		await page.goto(profile);
		await expect(page.getByText("This profile is private.")).toBeVisible();
		await expect(page.getByRole("link", { name: item.title })).toBeHidden();
		await expectAccessible(page);
	});

	test("follow a member and see their ratings in /feed", async ({ page, request, baseURL }) => {
		const item = await createItem();
		const followee = await createMember("followee");
		await insertRating(followee.id, item.id, [50, 50, 50, 50, 50, 50]);
		const account = await createVerifiedAccount(request, baseURL ?? "", "feed");
		await logIn(page, account);
		await expectLoggedIn(page);

		await page.goto("/feed");
		await expect(page.getByRole("heading", { level: 1, name: "Feed" })).toBeVisible();
		await expect(page.getByText("Your feed is empty. Follow people to see their ratings here.")).toBeVisible();

		await page.goto(`/u/${followee.nickname}`);
		const follow = page.getByRole("button", { name: "Follow", exact: true });
		await follow.click();
		const unfollow = page.getByRole("button", { name: "Unfollow", exact: true });
		await expect(unfollow).toHaveAttribute("aria-pressed", "true");
		await expect(unfollow).toBeEnabled();
		await expect(page.getByText(/^1 follower · 0 following$/)).toBeVisible();

		await page.goto("/feed");
		const entry = page.getByRole("listitem").filter({ hasText: followee.nickname });
		await expect(entry.getByRole("link", { name: item.title })).toBeVisible();
		await expect(entry.getByRole("img", { name: "Score 50 out of 100" })).toBeVisible();
		await expectAccessible(page);
	});

	test("/feed requires login", async ({ page }) => {
		await page.goto("/feed");
		await expect(page).toHaveURL(/\/login\?next=%2Ffeed/);
	});
});
