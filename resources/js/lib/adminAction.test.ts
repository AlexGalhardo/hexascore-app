import { afterEach, expect, test, vi } from "vitest";

import { postAction } from "./adminAction";

afterEach(() => vi.unstubAllGlobals());

const respond = (status: number, body: unknown) =>
	vi.stubGlobal(
		"fetch",
		vi.fn().mockResolvedValue({
			ok: status < 400,
			status,
			statusText: "Err",
			json: async () => body,
		}),
	);

test("a success is ok and sends a form-encoded body without empty or false fields", async () => {
	respond(200, {});
	const out = await postAction("/admin/x", { a: "1", b: false, c: undefined, d: true, e: 5 });
	expect(out).toEqual({ ok: true });
	const init = (fetch as unknown as ReturnType<typeof vi.fn>).mock.calls[0][1];
	const body = init.body as URLSearchParams;
	expect([...body.entries()]).toEqual([
		["a", "1"],
		["d", "on"],
		["e", "5"],
	]);
});

test("a problem detail becomes the message", async () => {
	respond(409, { detail: "A run of this job is already active." });
	expect(await postAction("/admin/x", {})).toEqual({
		ok: false,
		message: "A run of this job is already active.",
	});
});

test("validation errors are joined", async () => {
	respond(422, { errors: { reason: ["Choose a reject reason."], ids: ["Invalid form."] } });
	expect(await postAction("/admin/x", {})).toEqual({
		ok: false,
		message: "Choose a reject reason. Invalid form.",
	});
});

test("a body without a message falls back to the status line", async () => {
	respond(500, null);
	expect(await postAction("/admin/x", {})).toEqual({ ok: false, message: "500 Err" });
});
