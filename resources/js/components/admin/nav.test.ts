import { expect, test } from "vitest";

import { activeHref, NAV } from "./nav";

test("the longest matching href is the active one", () => {
	expect(activeHref("/admin/ingest")).toBe("/admin/ingest");
	expect(activeHref("/admin/ingest/reviews")).toBe("/admin/ingest/reviews");
	expect(activeHref("/admin/ingest/reviews/42")).toBe("/admin/ingest/reviews");
	expect(activeHref("/admin/overview")).toBe("/admin/overview");
});

test("an unknown path matches nothing", () => {
	expect(activeHref("/admin/unknown")).toBeNull();
	expect(activeHref("/profile")).toBeNull();
});

test("every href is unique and every item has a key", () => {
	const hrefs = NAV.flatMap((g) => g.items.map((i) => i.href));
	expect(new Set(hrefs).size).toBe(hrefs.length);
	expect(NAV.every((g) => g.items.every((i) => i.key.length > 0))).toBe(true);
});
