import { describe, expect, mock, test } from "bun:test";
import { type ContactDeps, type ContactMessage, submitContact } from "./contact";

const MESSAGE = "This is a long enough message for the contact form to accept.";

function form(values: Record<string, string>): FormData {
	const data = new FormData();
	for (const [k, v] of Object.entries(values)) data.set(k, v);
	return data;
}

function deps(overrides: Partial<ContactDeps> = {}): ContactDeps & { stored: ContactMessage[] } {
	const stored: ContactMessage[] = [];
	return {
		stored,
		user: null,
		ip: "203.0.113.7",
		isBot: async () => false,
		allow: async () => true,
		store: async (m) => {
			stored.push(m);
		},
		notify: async () => {},
		...overrides,
	};
}

describe("submitContact", () => {
	test("anonymous visitor: uses the submitted name and email", async () => {
		const d = deps();
		const result = await submitContact(
			form({ name: "Ana", email: "ana@example.com", subject: "question", message: MESSAGE }),
			d,
		);
		expect(result).toEqual({ ok: true });
		expect(d.stored[0]).toMatchObject({ name: "Ana", email: "ana@example.com" });
	});

	test("logged-in user: session overrides name and email from the body", async () => {
		const store = mock(async (_m: ContactMessage, _userId: string | null) => {});
		const d = deps({ user: { id: "u1", name: "Real Name", email: "real@example.com" }, store });
		const result = await submitContact(
			form({ name: "Spoofed", email: "victim@example.com", subject: "bug", message: MESSAGE }),
			d,
		);
		expect(result.ok).toBe(true);
		expect(store).toHaveBeenCalledWith(
			expect.objectContaining({ name: "Real Name", email: "real@example.com" }),
			"u1",
		);
	});

	test("rate-limit key is per user when logged in, per IP otherwise", async () => {
		const keys: string[] = [];
		const allow = async (key: string) => {
			keys.push(key);
			return true;
		};
		const body = { name: "A", email: "a@example.com", subject: "feedback", message: MESSAGE };
		await submitContact(form(body), deps({ allow }));
		await submitContact(form(body), deps({ allow, user: { id: "u1", name: "A", email: "a@example.com" } }));
		expect(keys).toEqual(["contact:ip:203.0.113.7", "contact:user:u1"]);
	});

	test.each([
		[{ name: "A", email: "not-an-email", subject: "question", message: MESSAGE }, ["email"]],
		[{ name: "A", email: "a@example.com", subject: "spam", message: MESSAGE }, ["subject"]],
		[{ name: "A", email: "a@example.com", subject: "question", message: "too short" }, ["message"]],
		[{ name: "A", email: "a@example.com", subject: "question", message: "x".repeat(513) }, ["message"]],
	])("rejects invalid input %#", async (body, fields) => {
		const d = deps();
		expect(await submitContact(form(body), d)).toEqual({ ok: false, error: "invalid", fields });
		expect(d.stored).toHaveLength(0);
	});

	test("blocks bots and rate-limited senders before storing", async () => {
		const body = { name: "A", email: "a@example.com", subject: "question", message: MESSAGE };
		const bot = deps({ isBot: async () => true });
		expect(await submitContact(form(body), bot)).toEqual({ ok: false, error: "bot", fields: undefined });
		const limited = deps({ allow: async () => false });
		expect(await submitContact(form(body), limited)).toEqual({
			ok: false,
			error: "rateLimited",
			fields: undefined,
		});
		expect([...bot.stored, ...limited.stored]).toHaveLength(0);
	});

	test("a failed notification still reports success (message is stored)", async () => {
		const d = deps({
			notify: async () => {
				throw new Error("smtp down");
			},
		});
		const body = { name: "A", email: "a@example.com", subject: "question", message: MESSAGE };
		expect(await submitContact(form(body), d)).toEqual({ ok: true });
		expect(d.stored).toHaveLength(1);
	});
});
