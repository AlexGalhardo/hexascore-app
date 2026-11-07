import type { Route } from "@playwright/test";
import { expect, test } from "../helpers/fixtures";

// Production base (APP_ENV=production) of the vukem-email subscribe endpoint, see app/http/middleware/newsletter.go.
const ENDPOINT = "https://email.hexascore.app/subscribe/hexascore";

// The browser sends a CORS preflight before the cross-origin JSON POST, so the mock answers it too.
async function mockSubscribe(
	route: Route,
	status: number,
	seen: { body?: unknown },
): Promise<void> {
	const cors = {
		"access-control-allow-origin": "*",
		"access-control-allow-methods": "POST, OPTIONS",
		"access-control-allow-headers": "content-type, accept",
	};
	if (route.request().method() === "OPTIONS") {
		await route.fulfill({ status: 204, headers: cors });
		return;
	}
	seen.body = route.request().postDataJSON();
	await route.fulfill({ status, headers: cors, contentType: "application/json", body: "{}" });
}

test.describe("footer newsletter", () => {
	for (const [status, copyKey] of [
		[202, "footer.newsletterDone"],
		[429, "footer.newsletterRateLimited"],
	] as const) {
		test(`answer ${status} shows its message, in the footer, with the address posted`, async ({
			page,
			prefix,
			t,
		}) => {
			const seen: { body?: unknown } = {};
			await page.route(ENDPOINT, (route) => mockSubscribe(route, status, seen));

			await page.goto(`${prefix}/`);
			// Typing before the page settles loses the value (suspected app bug, see COVERAGE.md "Found bugs"), so wait first.
			await page.waitForLoadState("networkidle");
			const form = page.locator("footer form").filter({
				has: page.locator('input[name="email"]'),
			});
			await form.locator('input[name="email"]').fill("reader@example.com");
			expect(await form.locator('input[name="email"]').inputValue()).toBe(
				"reader@example.com",
			);
			await form.getByRole("button", { name: t("footer.newsletterSubmit") }).click();

			await expect(form.getByText(t(copyKey))).toBeVisible();
			expect(seen.body).toEqual({ email: "reader@example.com", website: "" });
		});
	}
});
