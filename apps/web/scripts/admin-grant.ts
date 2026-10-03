// Usage: bun run admin:grant <email>   (stop `bun run dev` first when using PGlite: one process per data dir)
import { eq, sql } from "drizzle-orm";
import { z } from "zod";
import { db } from "../lib/db/client";
import { user } from "../lib/db/schema";

const email = z.email().safeParse(process.argv[2]?.trim().toLowerCase());
if (!email.success) {
	console.error("Usage: bun run admin:grant <email>");
	process.exit(1);
}

try {
	const rows = await db
		.update(user)
		.set({ role: "admin" })
		.where(eq(sql`lower(${user.email})`, email.data))
		.returning({ id: user.id, username: user.username });
	if (!rows.length) {
		console.error(`[admin] no user with email ${email.data}`);
		process.exit(1);
	}
	console.info(`[admin] granted admin to ${rows[0]?.username ?? rows[0]?.id} — sign in again if a session is open`);
	process.exit(0);
} catch (error) {
	console.error("[admin] grant failed:", error);
	process.exit(1);
}
