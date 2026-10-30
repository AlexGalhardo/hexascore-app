import { expect, test } from "bun:test";

import { ApiError, createClient } from "./api.ts";

test("clientErrors posts the body to /app/v1/client-errors", async () => {
	const body = { message: "boom" };
	let method = "";
	let url = "";
	let sent: string | undefined;
	const fetchImpl: typeof fetch = async (input, init) => {
		url =
			typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
		method = init?.method ?? "GET";
		sent = typeof init?.body === "string" ? init.body : undefined;
		return new Response(JSON.stringify({ ok: true }), {
			status: 200,
			headers: { "Content-Type": "application/json" },
		});
	};
	const client = createClient({ baseUrl: "http://app.test", fetch: fetchImpl });

	await client.clientErrors({ body });

	expect(method).toBe("POST");
	expect(url).toBe("http://app.test/app/v1/client-errors");
	expect(sent).toBe(JSON.stringify(body));
});

test("clientErrors throws ApiError when the response is not 2xx", async () => {
	const fetchImpl: typeof fetch = async () =>
		new Response(JSON.stringify({ title: "Nope", detail: "nope", status: 422 }), {
			status: 422,
			headers: { "Content-Type": "application/json" },
		});
	const client = createClient({ fetch: fetchImpl });
	let caught: unknown;
	try {
		await client.clientErrors();
	} catch (err) {
		caught = err;
	}

	expect(caught).toBeInstanceOf(ApiError);
	if (!(caught instanceof ApiError)) {
		throw new Error("expected ApiError");
	}
	expect(caught.status).toBe(422);
	expect(caught.problem?.detail).toBe("nope");
});
