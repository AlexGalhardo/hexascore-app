import { test } from "../helpers/fixtures";

// Milestone 0.3 (ratings and social) is being built in parallel. These placeholders keep the gaps visible in every
// report; replace each `fixme` with the real flow when its feature lands on `dev`.
test.describe("TODO 0.3 — ratings and social", () => {
	test.fixme("rating form: rate an item in 6 categories, edit it, metascore appears at the 5th rating", () => {});
	test.fixme("public profile /u/[nickname] shows ratings; private profile hides them", () => {});
	test.fixme("follow a user and see their ratings in /feed", () => {});
	test.fixme("item status (want / in progress / done / dropped) and curated lists", () => {});
	test.fixme("report content; admin resolves the report in /admin", () => {});
	test.fixme("/admin requires role=admin (403/redirect for regular users)", () => {});
});
