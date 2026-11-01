import { expect, test } from "vitest";

import { absoluteUrl } from "./Seo";

test("makes a share-card path absolute against the canonical origin", () => {
	expect(absoluteUrl("/og/items/1.png", "https://hexascore.app/games/hades")).toBe(
		"https://hexascore.app/og/items/1.png",
	);
	expect(absoluteUrl("https://cdn.example/c.webp", "https://hexascore.app/")).toBe(
		"https://cdn.example/c.webp",
	);
	expect(absoluteUrl("/og/items/1.png", undefined)).toBe("/og/items/1.png");
});
