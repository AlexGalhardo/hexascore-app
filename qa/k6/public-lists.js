// k6: public list pages (SSR) at 30 rps, p95 <= 400 ms, errors < 0.5 % (SDD 14 NFR-14-1, §9).
// Needs LIST_PATHS: comma-separated /u/<nickname>/lists/<id> paths of public lists, seeded by qa/k6/seed.sh.
// Env: BASE_URL, LIST_PATHS, DURATION (60s). Run:
// docker run --rm --network host -e BASE_URL -e LIST_PATHS -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/public-lists.js
import { check, fail } from "k6";
import { scenario } from "k6/execution";
import http from "k6/http";

// Distinct client address per request so the per-IP limits do not mask latency.
const fakeIP = (n) => `10.${(n >> 8) % 250}.${n % 250}.${1 + ((n * 7) % 250)}`;
const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");
const PATHS = (__ENV.LIST_PATHS || "").split(",").filter(Boolean);

export const options = {
	scenarios: {
		lists: {
			executor: "constant-arrival-rate",
			rate: 30,
			timeUnit: "1s",
			duration: __ENV.DURATION || "60s",
			preAllocatedVUs: 30,
			maxVUs: 100,
		},
	},
	thresholds: {
		http_req_duration: ["p(95)<400"],
		http_req_failed: ["rate<0.005"],
		checks: ["rate>0.995"],
	},
};

export function setup() {
	if (PATHS.length === 0) fail("set LIST_PATHS (qa/k6/seed.sh)");
}

export default function () {
	const res = http.get(`${BASE}${PATHS[scenario.iterationInTest % PATHS.length]}`, {
		headers: { "X-Forwarded-For": fakeIP(scenario.iterationInTest) },
	});
	check(res, { "list page: 200": (r) => r.status === 200 });
}
