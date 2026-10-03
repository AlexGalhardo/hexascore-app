import { expect, test } from "bun:test";
import { z } from "zod";
import { decodeCursor, encodeCursor } from "./cursor";

test("cursor round-trips and rejects garbage", () => {
	const schema = z.object({ a: z.number() });
	expect(decodeCursor(encodeCursor({ a: 1 }), schema)).toEqual({ a: 1 });
	expect(decodeCursor(encodeCursor({ a: "1" }), schema)).toBeNull();
	expect(decodeCursor("eyJh", schema)).toBeNull();
	expect(decodeCursor("+/=", schema)).toBeNull();
});
