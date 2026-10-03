import { createVerifiedAccount, expectLoggedIn, hasAuthSecret, logIn } from "../helpers/auth";
import { expect, expectAccessible, test } from "../helpers/fixtures";

test.describe("contact", () => {
	test("counter follows the message and the form validates", async ({ page }) => {
		await page.goto("/contact");
		const counter = page.locator("#contact-message-count");
		await expect(counter).toContainText("0/");
		await page.getByLabel("Message").fill("Olá 👋 ünïcödé");
		await expect(counter).toContainText(/1\d\//);
		await expectAccessible(page);

		await page.getByLabel("Name").fill("QA");
		await page.getByLabel("Email").fill("not-an-email");
		await page.getByLabel("Message").fill("short");
		await page.getByRole("button", { name: "Send message" }).click();
		await expect(page.getByText("Message sent.")).toBeHidden();
	});

	test("message length is capped by the field", async ({ page }) => {
		await page.goto("/contact");
		const message = page.getByLabel("Message");
		const max = Number(await message.getAttribute("maxlength"));
		expect(max).toBeGreaterThan(0);
		await message.fill("x".repeat(max + 50));
		expect((await message.inputValue()).length).toBe(max);
	});
});

test.describe("contact while logged in @auth", () => {
	test.skip(!hasAuthSecret(), "needs E2E_BETTER_AUTH_SECRET to create a verified account");

	test("sends a message with identity from the session", async ({ page, request, baseURL }) => {
		const account = await createVerifiedAccount(request, baseURL ?? "", "contact");
		await logIn(page, account);
		await expectLoggedIn(page);
		await page.goto("/contact");
		await expect(page.getByText("Sent from your account.")).toBeVisible();
		await page.getByLabel("Subject").selectOption({ index: 1 });
		await page
			.getByLabel("Message")
			.fill("E2E: <script>alert(1)</script> ' OR 1=1 -- this is a long enough message.");
		const submit = page.getByRole("button", { name: "Send message" });
		await submit.click();
		await expect(page.getByText(/Message sent/)).toBeVisible();
	});
});
