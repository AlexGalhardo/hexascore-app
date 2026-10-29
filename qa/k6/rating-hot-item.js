// k6: 50 concurrent raters on one hot item (SDD 02 NFR-02-1, open issue 02-3).
// Env: BASE_URL, ITEM_ID (item with ~10,000 ratings), TOKENS (comma-separated API tokens with ratings:write,
// at least 50, one per VU). Run: docker run --rm --network host -e BASE_URL -e ITEM_ID -e TOKENS
// -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/rating-hot-item.js
import { check, fail } from "k6";
import http from "k6/http";

const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");
const TOKENS = (__ENV.TOKENS || "").split(",").filter(Boolean);
const ITEM = __ENV.ITEM_ID || "";

export const options = {
	scenarios: {
		raters: { executor: "constant-vus", vus: 50, duration: "60s" },
	},
	thresholds: {
		"http_req_duration{name:rating}": ["p(95)<60"],
		checks: ["rate>0.99"],
	},
};

export function setup() {
	if (!ITEM || TOKENS.length < 50) fail("set ITEM_ID and at least 50 TOKENS");
}

const keys = ["gameplay", "fun", "sound", "technical", "visual", "storytelling"];

export default function () {
	const token = TOKENS[(__VU - 1) % TOKENS.length];
	const headers = { Authorization: `Bearer ${token}`, "Content-Type": "application/json" };
	const scores = {};
	for (const k of keys) scores[k] = 1 + Math.floor(Math.random() * 10) / 2;
	const res = http.put(`${BASE}/app/v1/items/${ITEM}/rating`, JSON.stringify({ scores }), {
		headers,
		tags: { name: "rating" },
	});
	check(res, { "rating saved": (r) => r.status === 200 });
}
