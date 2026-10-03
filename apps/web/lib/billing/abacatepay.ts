import { createHash, createHmac, timingSafeEqual } from "node:crypto";
import { z } from "zod";
import { env } from "../env";
import { fetchWithTimeout } from "../http";

// API v2: https://docs.abacatepay.com/pages/transparents/create · /check · /simulate-payment
const BASE_URL = "https://api.abacatepay.com/v2";

/**
 * Public HMAC key published by AbacatePay (https://docs.abacatepay.com/pages/webhooks): every webhook carries
 * `X-Webhook-Signature` = base64(HMAC-SHA256(rawBody, this key)). It is the same for every merchant, so the
 * per-webhook `?webhookSecret=` (ABACATEPAY_WEBHOOK_SECRET) is what proves the call targets our endpoint.
 */
export const ABACATEPAY_PUBLIC_HMAC_KEY =
	"t9dXRhHHo3yDEj5pVDYz0frf7q6bMKyMRmxxCPIPp3RCplBfXRxqlC6ZpiWmOqj4L63qEaeUOtrCI8P0VMUgo6iIga2ri9ogaHFs0WIIywSMg0q7RmBfybe1E5XJcfC4IW3alNqym0tXoAKkzvfEjZxV6bE0oG2zJrNNYmUCKZyV0KZ3JS8Votf9EAWWYdiDkMkpbMdPggfh1EqHlVkMiTady6jOR3hyzGEHrIz2Ret0xHKMbiqkr9HS1JhNHDX9";

export const PIX_STATUSES = [
	"PENDING",
	"PAID",
	"EXPIRED",
	"CANCELLED",
	"REFUNDED",
	"REDEEMED",
	"APPROVED",
	"FAILED",
	"UNDER_DISPUTE",
] as const;
export type PixProviderStatus = (typeof PIX_STATUSES)[number];

const pixQrCodeSchema = z.object({
	id: z.string().min(1),
	amount: z.number().int(),
	status: z.enum(PIX_STATUSES),
	brCode: z.string().min(1),
	brCodeBase64: z.string().min(1),
	expiresAt: z.coerce.date(),
});
export type PixQrCode = z.infer<typeof pixQrCodeSchema>;

const checkSchema = z.object({ id: z.string().optional(), status: z.enum(PIX_STATUSES) });

const envelopeSchema = z.object({ data: z.unknown(), error: z.unknown().optional() });

async function call<T extends z.ZodType>(path: string, init: RequestInit, schema: T): Promise<z.infer<T>> {
	if (!env.ABACATEPAY_API_KEY) throw new Error("AbacatePay is not configured (ABACATEPAY_API_KEY missing).");
	const response = await fetchWithTimeout(`${BASE_URL}${path}`, {
		...init,
		headers: { Authorization: `Bearer ${env.ABACATEPAY_API_KEY}`, "Content-Type": "application/json" },
	});
	const body = envelopeSchema.safeParse(await response.json());
	if (!body.success || body.data.data == null) {
		throw new Error(
			`AbacatePay ${path} failed: ${JSON.stringify(body.success ? body.data.error : body.error.message)}`,
		);
	}
	const parsed = schema.safeParse(body.data.data);
	if (!parsed.success) throw new Error(`Unexpected AbacatePay response from ${path}: ${parsed.error.message}`);
	return parsed.data;
}

export async function createPixQrCode(input: {
	amountCents: number;
	description: string;
	expiresInSeconds: number;
	metadata: Record<string, string>;
}): Promise<PixQrCode> {
	return call(
		"/transparents/create",
		{
			method: "POST",
			body: JSON.stringify({
				method: "PIX",
				data: {
					amount: input.amountCents,
					description: input.description.slice(0, 500),
					expiresIn: input.expiresInSeconds,
					metadata: input.metadata,
				},
			}),
		},
		pixQrCodeSchema,
	);
}

export async function checkPixQrCode(id: string): Promise<PixProviderStatus> {
	const data = await call(`/transparents/check?id=${encodeURIComponent(id)}`, { method: "GET" }, checkSchema);
	return data.status;
}

/** Dev keys only (production keys get an error): marks the charge as paid and fires the webhook, like a real payment. */
export async function simulatePixPayment(id: string): Promise<PixProviderStatus> {
	const data = await call(
		`/transparents/simulate-payment?id=${encodeURIComponent(id)}`,
		{ method: "POST", body: "{}" },
		checkSchema,
	);
	return data.status;
}

/** Constant-time string compare that does not leak the length (both sides hashed first). */
export function safeEqual(a: string, b: string): boolean {
	const ha = createHash("sha256").update(a).digest();
	const hb = createHash("sha256").update(b).digest();
	return timingSafeEqual(ha, hb) && a.length === b.length;
}

export function signAbacatePayload(rawBody: string, key = ABACATEPAY_PUBLIC_HMAC_KEY): string {
	return createHmac("sha256", key).update(Buffer.from(rawBody, "utf8")).digest("base64");
}

export function verifyAbacateSignature(
	rawBody: string,
	signature: string | null,
	key = ABACATEPAY_PUBLIC_HMAC_KEY,
): boolean {
	if (!signature) return false;
	return safeEqual(signAbacatePayload(rawBody, key), signature.trim());
}

// AbacatePay asks integrators not to validate whole payloads (new fields must not break us): only what we read.
export const abacateWebhookSchema = z.looseObject({
	id: z.string().min(1),
	event: z.string().min(1),
	devMode: z.boolean().optional(),
	data: z.looseObject({
		id: z.string().optional(),
		amount: z.number().int().optional(),
		status: z.string().optional(),
		transparent: z.looseObject({ id: z.string(), amount: z.number().int().optional() }).optional(),
		pixQrCode: z.looseObject({ id: z.string(), amount: z.number().int().optional() }).optional(),
	}),
});
export type AbacateWebhook = z.infer<typeof abacateWebhookSchema>;

export function chargeRefFromWebhook(payload: AbacateWebhook): { id: string; amount: number | null } | null {
	const d = payload.data;
	const nested = d.transparent ?? d.pixQrCode;
	const id = nested?.id ?? d.id;
	if (!id) return null;
	return { id, amount: nested?.amount ?? d.amount ?? null };
}
