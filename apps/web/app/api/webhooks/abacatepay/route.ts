import { getStore } from "@/lib/api/quota";
import { handleAbacateWebhook } from "@/lib/billing/pix";
import { db } from "@/lib/db/client";
import { env } from "@/lib/env";

export async function POST(request: Request): Promise<Response> {
	if (!env.ABACATEPAY_WEBHOOK_SECRET) {
		console.error("[pix] webhook received but ABACATEPAY_WEBHOOK_SECRET is not set: returning 503");
		return Response.json({ error: "pix webhooks not configured" }, { status: 503 });
	}
	// HMAC is computed over the raw body: read text, never a re-serialized JSON.
	const rawBody = await request.text();
	const result = await handleAbacateWebhook(
		rawBody,
		{
			webhookSecret: new URL(request.url).searchParams.get("webhookSecret"),
			signature: request.headers.get("x-webhook-signature"),
		},
		{ db, store: getStore(), secret: env.ABACATEPAY_WEBHOOK_SECRET },
	);
	return Response.json(result.body, { status: result.status });
}
