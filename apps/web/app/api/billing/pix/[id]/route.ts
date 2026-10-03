import { and, eq } from "drizzle-orm";
import { z } from "zod";
import { getStore } from "@/lib/api/quota";
import { checkPixQrCode } from "@/lib/billing/abacatepay";
import { settlePixCharge } from "@/lib/billing/pix";
import { db } from "@/lib/db/client";
import { pixCharges } from "@/lib/db/schema";
import { features } from "@/lib/env";
import { getSessionUser } from "@/lib/session";

const paramsSchema = z.object({ id: z.uuid() });

export type PixStatusResponse = { status: "pending" | "paid" | "expired" | "failed" };

// Polled by the checkout page every few seconds; also re-checks AbacatePay so a lost webhook never strands a payment.
export async function GET(_request: Request, ctx: { params: Promise<{ id: string }> }): Promise<Response> {
	const user = await getSessionUser();
	if (!user) return Response.json({ error: "unauthenticated" }, { status: 401 });
	const params = paramsSchema.safeParse(await ctx.params);
	if (!params.success) return Response.json({ error: "not found" }, { status: 404 });

	try {
		const [charge] = await db
			.select()
			.from(pixCharges)
			.where(and(eq(pixCharges.id, params.data.id), eq(pixCharges.userId, user.id)));
		if (!charge) return Response.json({ error: "not found" }, { status: 404 });

		let status = charge.status;
		if (status === "pending" && features.pix) {
			try {
				const remote = await checkPixQrCode(charge.providerRef);
				if (remote === "PAID") {
					await settlePixCharge(db, getStore(), charge.id);
					status = "paid";
				} else if (remote === "EXPIRED" || remote === "CANCELLED" || remote === "FAILED") {
					status = remote === "FAILED" ? "failed" : "expired";
					await db
						.update(pixCharges)
						.set({ status })
						.where(and(eq(pixCharges.id, charge.id), eq(pixCharges.status, "pending")));
				}
			} catch (error) {
				console.warn(`[pix] status re-check failed for ${charge.id}`, error);
			}
		}
		return Response.json({ status } satisfies PixStatusResponse, { headers: { "Cache-Control": "no-store" } });
	} catch (error) {
		console.error("[pix] status lookup failed", error);
		return Response.json({ error: "server error" }, { status: 500 });
	}
}
