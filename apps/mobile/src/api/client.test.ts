import { ApiError } from "../../../../packages/app-client/src/api";
import { createAppClient, uuidv7 } from "./client";
import { describeError } from "./problems";
import { compareVersions } from "./version";

type Call = { url: string; method: string; headers: Headers };

function stub(
	status: number,
	body: unknown,
	extra: Record<string, string> = {},
): { fetch: typeof fetch; calls: Call[] } {
	const calls: Call[] = [];
	const impl = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
		calls.push({
			url: String(input),
			method: init?.method ?? "GET",
			headers: new Headers(init?.headers),
		});
		return new Response(JSON.stringify(body), {
			status,
			headers: { "Content-Type": "application/json", ...extra },
		});
	};
	return { fetch: impl as typeof fetch, calls };
}

const base = { baseUrl: "https://api.test", version: "1.2.0", locale: (): "pt-BR" => "pt-BR" };

test("reads send Authorization, X-App-Version and Accept-Language but no Idempotency-Key", async () => {
	const { fetch: f, calls } = stub(200, {});
	const api = createAppClient({ ...base, fetch: f, token: () => "hx_app_abc" });
	await api.read.meShow();
	expect(calls[0]?.url).toBe("https://api.test/app/v1/me");
	expect(calls[0]?.headers.get("Authorization")).toBe("Bearer hx_app_abc");
	expect(calls[0]?.headers.get("X-App-Version")).toBe("1.2.0");
	expect(calls[0]?.headers.get("Accept-Language")).toBe("pt-BR");
	expect(calls[0]?.headers.has("Idempotency-Key")).toBe(false);
});

test("no token means no Authorization header", async () => {
	const { fetch: f, calls } = stub(200, {});
	await createAppClient({ ...base, fetch: f }).read.meShow();
	expect(calls[0]?.headers.has("Authorization")).toBe(false);
});

test("writes carry the caller's Idempotency-Key", async () => {
	const { fetch: f, calls } = stub(200, {});
	const api = createAppClient({ ...base, fetch: f });
	const key = api.newIdempotencyKey();
	await api.write(key, (c) => c.ratingsPut({ params: { id: "1" }, body: { score: 4 } }));
	expect(calls[0]?.method).toBe("PUT");
	expect(calls[0]?.headers.get("Idempotency-Key")).toBe(key);
});

test("a failed write carries Retry-After on the ApiError", async () => {
	const { fetch: f } = stub(
		429,
		{ status: 429, code: "item-ratings-limited" },
		{ "Retry-After": "120", "Content-Type": "application/problem+json" },
	);
	const api = createAppClient({ ...base, fetch: f });
	const error: unknown = await api
		.write(api.newIdempotencyKey(), (c) => c.ratingsPut({ params: { id: "1" }, body: {} }))
		.catch((e: unknown) => e);
	expect(error).toBeInstanceOf(ApiError);
	expect((error as { retryAfter?: unknown }).retryAfter).toBe(120);
});

test("a write without an idempotency key does not type-check", () => {
	const api = createAppClient({ ...base, fetch: stub(200, {}).fetch });
	// Compile-time only: the function is never called.
	const typeOnly = (): void => {
		// @ts-expect-error write methods are not on the read client
		void api.read.ratingsPut;
		// @ts-expect-error a plain string is not an IdempotencyKey
		void api.write("not-branded", (c) => c.meUpdate());
		// @ts-expect-error the key is required
		void api.write((c) => c.meUpdate());
	};
	expect(typeof typeOnly).toBe("function");
});

test("uuidv7 is version 7, variant 10xx, time ordered", () => {
	const bytes = new Uint8Array(16).fill(0xff);
	const a = uuidv7(1_700_000_000_000, () => bytes);
	const b = uuidv7(1_700_000_000_001, () => bytes);
	expect(a).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
	expect(a < b).toBe(true);
});

test("compareVersions orders semver numerically", () => {
	expect(compareVersions("1.10.0", "1.9.9")).toBeGreaterThan(0);
	expect(compareVersions("1.2.0", "1.2.0")).toBe(0);
	expect(compareVersions("0.0.0", "0.0.1")).toBeLessThan(0);
});

test("X-App-Min-Version above the running version sets upgradeRequired", async () => {
	const { fetch: f } = stub(200, {}, { "X-App-Min-Version": "1.3.0" });
	const seen: boolean[] = [];
	const api = createAppClient({ ...base, fetch: f, onUpgradeRequired: () => seen.push(true) });
	expect(api.upgradeRequired()).toBe(false);
	await api.read.meShow();
	expect(api.upgradeRequired()).toBe(true);
	expect(seen).toEqual([true]);
});

test("a min version at or below the running one does not", async () => {
	const { fetch: f } = stub(200, {}, { "X-App-Min-Version": "1.2.0" });
	const api = createAppClient({ ...base, fetch: f });
	await api.read.meShow();
	expect(api.upgradeRequired()).toBe(false);
});

test("a 426 sets upgradeRequired even without the header", async () => {
	const { fetch: f } = stub(426, { code: "app-upgrade-required", status: 426 });
	const api = createAppClient({ ...base, fetch: f });
	await expect(api.read.meShow()).rejects.toBeInstanceOf(ApiError);
	expect(api.upgradeRequired()).toBe(true);
});

test("a known code maps to errors.<code>", () => {
	const err = new ApiError(403, { code: "account-banned", request_id: "r1" } as never);
	expect(describeError(err, (key) => key === "errors.account-banned")).toEqual({
		key: "errors.account-banned",
		requestId: "r1",
	});
});

test("an unknown code maps to errors.unexpected with the request_id", () => {
	const err = new ApiError(500, { code: "never-heard", request_id: "r2" } as never);
	expect(describeError(err, () => false)).toEqual({ key: "errors.unexpected", requestId: "r2" });
});

test("a non-API failure maps to errors.unexpected without a request_id", () => {
	expect(describeError(new TypeError("Network request failed"), () => true)).toEqual({
		key: "errors.unexpected",
		requestId: undefined,
	});
});
