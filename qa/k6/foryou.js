// k6: GET /app/v1/me/for-you, p(95) <= 400 ms on a miss and <= 30 ms on a hit (SDD 08 NFR-08-6).
// Env: BASE_URL, TOKENS (comma-separated app tokens with catalog:read, each of a user with >= 3 rated items of TYPE; the
// first call per user is the miss, later calls are hits), TYPE (game), VUS (5), DURATION (60s).
// docker run --rm --network host -e BASE_URL -e TOKENS -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/foryou.js
import { check, fail } from "k6";
import http from "k6/http";
import { Trend } from "k6/metrics";

const BASE = `${(__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "")}/app/v1`;
const TOKENS = (__ENV.TOKENS || "").split(",").filter(Boolean);
const TYPE = __ENV.TYPE || "game";
const miss = new Trend("foryou_miss", true);
const hit = new Trend("foryou_hit", true);

export const options = {
	vus: Number(__ENV.VUS || 5),
	duration: __ENV.DURATION || "60s",
	thresholds: { foryou_miss: ["p(95)<400"], foryou_hit: ["p(95)<30"], http_req_failed: ["rate<0.001"] },
};

export function setup() {
	if (TOKENS.length === 0) fail("TOKENS is required");
}

const warmed = {};

export default function () {
	const token = TOKENS[(__VU - 1) % TOKENS.length];
	const res = http.get(`${BASE}/me/for-you?type=${TYPE}`, { headers: { Authorization: `Bearer ${token}` } });
	check(res, { "200": (r) => r.status === 200 });
	(warmed[token] ? hit : miss).add(res.timings.duration);
	warmed[token] = true;
}
