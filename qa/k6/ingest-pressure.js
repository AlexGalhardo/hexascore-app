// k6: item-page p95 stays within 1.2x the idle p95 while a heavy ingestion run (Friday TMDB) is under way (SDD 06 §9, NFR-06-1).
// Run it twice against staging: once idle to read IDLE_P95_MS from http_req_duration p(95), once during the run with
// IDLE_P95_MS set. Env: BASE_URL, API_KEY (only to pick item slugs), IDLE_P95_MS (default 300), DURATION (5m). Run:
// docker run --rm --network host -e BASE_URL -e API_KEY -e IDLE_P95_MS -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/ingest-pressure.js
import { check, fail } from "k6";
import http from "k6/http";

const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");
const IDLE = Number(__ENV.IDLE_P95_MS || 300);
const fakeIP = (n) => `10.${(n >> 8) % 250}.${n % 250}.${1 + ((n * 7) % 250)}`;

export const options = {
	scenarios: {
		items: {
			executor: "constant-arrival-rate",
			rate: 10,
			timeUnit: "1s",
			duration: __ENV.DURATION || "5m",
			preAllocatedVUs: 10,
			maxVUs: 50,
		},
	},
	thresholds: { http_req_failed: ["rate<0.005"], http_req_duration: [`p(95)<${Math.round(IDLE * 1.2)}`] },
};

export function setup() {
	const key = (__ENV.API_KEY || "").split(",")[0];
	if (!key) fail("set API_KEY");
	const res = http.get(`${BASE}/api/v1/movies?limit=100`, { headers: { Authorization: `Bearer ${key}` } });
	if (res.status !== 200) fail(`movies list answered ${res.status}`);
	return { slugs: res.json("data").map((x) => x.slug) };
}

export default function (data) {
	const slug = data.slugs[Math.floor(Math.random() * data.slugs.length)];
	const res = http.get(`${BASE}/movies/${slug}`, { headers: { "X-Forwarded-For": fakeIP(__ITER) } });
	check(res, { "item page 200": (r) => r.status === 200 });
}
