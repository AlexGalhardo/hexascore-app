import { getStore } from "@/lib/api/quota";
import { getStripe } from "@/lib/billing/stripe";
import { handleStripeWebhook } from "@/lib/billing/stripe-webhook";
import { db } from "@/lib/db/client";
import { env } from "@/lib/env";

export async function POST(request: Request): Promise<Response> {
	const stripe = getStripe();
	// Checkout works without the webhook secret, but events cannot be trusted until it is set: refuse loudly.
	if (!stripe || !env.STRIPE_WEBHOOK_SECRET) {
		console.error(
			`[stripe] webhook received but ${stripe ? "STRIPE_WEBHOOK_SECRET" : "STRIPE_SECRET_KEY"} is not set: returning 503`,
		);
		return Response.json({ error: "stripe webhooks not configured" }, { status: 503 });
	}
	// The signature covers the exact bytes Stripe sent: read the raw text, never a re-serialized JSON.
	const rawBody = await request.text();
	const result = await handleStripeWebhook(rawBody, request.headers.get("stripe-signature"), {
		stripe,
		secret: env.STRIPE_WEBHOOK_SECRET,
		db,
		store: getStore(),
		retrieveSubscription: (id) => stripe.subscriptions.retrieve(id),
	});
	return Response.json(result.body, { status: result.status });
}
