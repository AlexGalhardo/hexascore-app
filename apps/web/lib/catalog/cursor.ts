import { z } from "zod";

// Opaque keyset cursor: base64url(JSON). Clients must treat it as a black box; the shape may change.
export function encodeCursor(value: unknown): string {
	return Buffer.from(JSON.stringify(value), "utf8").toString("base64url");
}

export function decodeCursor<T>(cursor: string, schema: z.ZodType<T>): T | null {
	if (cursor.length > 512 || !/^[A-Za-z0-9_-]+$/.test(cursor)) return null;
	try {
		const parsed = schema.safeParse(JSON.parse(Buffer.from(cursor, "base64url").toString("utf8")));
		return parsed.success ? parsed.data : null;
	} catch {
		return null;
	}
}

export class InvalidCursorError extends Error {
	constructor() {
		super("Invalid cursor");
		this.name = "InvalidCursorError";
	}
}

export const keysetValue = z.union([z.number(), z.string().max(64)]);
