import { expect, test } from "bun:test";
import { nicknameSchema } from "@/lib/validation";
import { generateItem, generateRating, generateUser, raterIndexes } from "./seed-data";

test("item generation is deterministic and independent of catalog size", () => {
	expect(generateItem("game", 7, 50)).toEqual(generateItem("game", 7, 50));
	generateItem("book", 3, 50);
	const again = generateItem("game", 7, 50);
	expect(again.wikidataQid).toBe("seed-game-000007");
	expect(again.translations[0]?.locale).toBe("en");
	expect(again.meta.trailers?.every((t) => /^FAKE[A-Za-z0-9]{7}$/.test(t.youtubeId)) ?? true).toBe(true);
});

test("seed users are valid, unique and use the reserved test domain", () => {
	const users = Array.from({ length: 200 }, (_, i) => generateUser(i));
	expect(new Set(users.map((u) => u.username)).size).toBe(200);
	for (const u of users) {
		expect(nicknameSchema.safeParse(u.username).success).toBe(true);
		expect(u.email.endsWith("@example.test")).toBe(true);
	}
});

test("ratings stay in range and raters are distinct", () => {
	const scores = generateRating({ wikidataQid: "seed-movie-000001", quality: 95 }, 3);
	expect(scores).toHaveLength(6);
	expect(scores.every((s) => s >= 0 && s <= 100)).toBe(true);
	const raters = raterIndexes(5, 40, 50);
	expect(new Set(raters).size).toBe(40);
});

test("refuses to run against production", () => {
	const result = Bun.spawnSync(["bun", "lib/db/seed.ts"], {
		cwd: `${import.meta.dir}/../..`,
		env: { ...process.env, VERCEL_ENV: "production" },
	});
	expect(result.exitCode).toBe(1);
	expect(result.stderr.toString()).toContain("Refusing to seed");
});
