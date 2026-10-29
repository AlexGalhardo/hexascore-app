import { open } from "../helpers/account";
import { expectNoSeriousA11yViolations } from "../helpers/axe";
import { expect, test } from "../helpers/fixtures";

const LONG_ENOUGH = "E2E: <script>alert(1)</script> ' OR 1=1 -- this is a long enough message.";

test.describe("contact", () => {
	test("the form is accessible and validates before sending", async ({ page, prefix, t }) => {
		await open(page, `${prefix}/contact`);
		await expectNoSeriousA11yViolations(page);

		await page.getByLabel(t("contact.name")).fill("QA");
		await page.getByLabel(t("contact.email")).fill("not-an-email");
		await page.getByLabel(t("contact.message")).fill("short");
		await page.getByRole("button", { name: t("contact.submit") }).click();
		await expect(page.getByText(t("contact.sent"))).toBeHidden();
	});

	test("message length is capped by the field", async ({ page, prefix, t }) => {
		await open(page, `${prefix}/contact`);
		const message = page.getByLabel(t("contact.message"));
		const max = Number(await message.getAttribute("maxlength"));
		expect(max).toBeGreaterThan(0);
		await message.fill("x".repeat(max + 50));
		expect((await message.inputValue()).length).toBe(max);
	});

	test("a valid message is sent and hostile markup stays inert", async ({
		page,
		prefix,
		t,
		any,
	}) => {
		await open(page, `${prefix}/contact`);
		await page.getByLabel(t("contact.name")).fill("QA");
		await page.getByLabel(t("contact.email")).fill("qa@example.test");
		await page.getByLabel(t("contact.subject")).selectOption("question");
		await page.getByLabel(t("contact.message")).fill(LONG_ENOUGH);
		// The server refuses a form submitted less than 3 s after it was served (bot guard).
		await page.waitForTimeout(3_500);
		await page.getByRole("button", { name: t("contact.submit") }).click();
		await expect(page.getByText(any("contact.sent"))).toBeVisible();
	});
});
