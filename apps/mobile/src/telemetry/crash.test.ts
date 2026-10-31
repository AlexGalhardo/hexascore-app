import { type ConsentStorage, readCrashConsent, writeCrashConsent } from "./consent";
import {
	buildCrashReport,
	type CrashContext,
	createCrashReporter,
	installGlobalHandler,
	STACK_MAX_BYTES,
} from "./crash";

function storage(): ConsentStorage & { data: Map<string, string> } {
	const data = new Map<string, string>();
	return {
		data,
		getItemSync: (key) => data.get(key) ?? null,
		setItem: async (key, value) => {
			data.set(key, value);
		},
	};
}

const context: CrashContext = { appVersion: "0.1.0", versionCode: 7, androidApi: 34 };
const flush = (): Promise<void> => new Promise((r) => setImmediate(r));

describe("consent", () => {
	test("is unset on first launch, then remembers the choice and can be revoked", async () => {
		const s = storage();
		expect(readCrashConsent(s)).toBe("unset");
		await writeCrashConsent(s, true);
		expect(readCrashConsent(s)).toBe("granted");
		await writeCrashConsent(s, false);
		expect(readCrashConsent(s)).toBe("denied");
	});

	test("an unknown stored value counts as unset (default off)", () => {
		const s = storage();
		s.data.set("hexascore.crashConsent", "maybe");
		expect(readCrashConsent(s)).toBe("unset");
	});
});

describe("reporter", () => {
	function setup(consent: "unset" | "granted" | "denied") {
		const sent: unknown[] = [];
		const report = createCrashReporter({
			consent: () => consent,
			send: async (body) => {
				sent.push(body);
			},
			context,
			screen: () => "/[type]/[slug]",
		});
		return { sent, report };
	}

	test("nothing is sent without consent", async () => {
		for (const state of ["unset", "denied"] as const) {
			const { sent, report } = setup(state);
			report(new Error("boom"));
			await flush();
			expect(sent).toEqual([]);
		}
	});

	test("with consent it posts exactly the SDD 10 §7.7 fields", async () => {
		const { sent, report } = setup("granted");
		report(new Error("boom"));
		await flush();
		expect(sent).toHaveLength(1);
		expect(Object.keys(sent[0] as object).sort()).toEqual(
			["androidApi", "appVersion", "message", "screen", "stack", "versionCode"].sort(),
		);
		expect(sent[0]).toMatchObject({
			appVersion: "0.1.0",
			versionCode: 7,
			androidApi: 34,
			screen: "/[type]/[slug]",
			message: "boom",
		});
	});

	test("a failing send never throws back into the app", async () => {
		const report = createCrashReporter({
			consent: () => "granted",
			send: async () => {
				throw new Error("offline");
			},
			context,
			screen: () => "/",
		});
		expect(() => report(new Error("boom"))).not.toThrow();
		await flush();
	});
});

describe("payload", () => {
	test("has no user id, token or e-mail", () => {
		const error = new Error("401 for ana@example.com with Bearer hx_app_s3cr3t");
		error.stack = `Error: token hx_app_s3cr3t user ana@example.com\n    at fn (app.js:1:2)`;
		const report = buildCrashReport(error, "/users/[username]", context);
		const text = JSON.stringify(report);
		expect(text).not.toMatch(/ana@example\.com|hx_app_s3cr3t|Bearer/);
		expect(Object.keys(report).join()).not.toMatch(/user|token|mail/i);
		expect(report.message).toContain("[email]");
		expect(report.stack).toContain("[token]");
	});

	test("the stack is truncated to 8 KB", () => {
		const error = new Error("big");
		error.stack = "é".repeat(10_000);
		const report = buildCrashReport(error, "/", context);
		expect(new TextEncoder().encode(report.stack).length).toBeLessThanOrEqual(STACK_MAX_BYTES);
		expect(report.stack.length).toBeGreaterThan(4000);
	});

	test("a non-Error throw still produces a message, and the screen drops query strings", () => {
		const report = buildCrashReport("plain", "/search?q=ana@example.com", context);
		expect(report.message).toBe("plain");
		expect(report.screen).toBe("/search");
	});
});

describe("global handler", () => {
	test("reports and then calls the previous handler; uninstall restores it", () => {
		let handler: (error: unknown, isFatal?: boolean) => void = () => {};
		const previous = jest.fn();
		handler = previous;
		const errorUtils = {
			getGlobalHandler: () => handler,
			setGlobalHandler: (next: typeof handler) => {
				handler = next;
			},
		};
		const report = jest.fn();
		const uninstall = installGlobalHandler(report, errorUtils);
		const error = new Error("fatal");
		handler(error, true);
		expect(report).toHaveBeenCalledWith(error);
		expect(previous).toHaveBeenCalledWith(error, true);
		uninstall();
		expect(handler).toBe(previous);
	});
});
