// k6: public API /api/v1 list and detail at 50 rps combined (SDD 05 NFR-05-1, §9). 60 s by default; DURATION=10m for the full NFR.
// Needs API_KEYS: comma-separated Business keys (50 rps each), seeded by qa/k6/seed.sh. Each request sends a distinct
// X-Forwarded-For so the 120/min per-IP limit does not mask latency (the app must trust the proxy, see docker-compose.yml).
// Env: BASE_URL, API_KEYS, DURATION, LIST_RPS (25), DETAIL_RPS (25). Run:
// docker run --rm --network host -e BASE_URL -e API_KEYS -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/api-v1.js
import { check, fail } from "k6";
import { scenario } from "k6/execution";
import http from "k6/http";

// Distinct client address per request so the per-IP limits do not mask latency.
const fakeIP = (n) => `10.${(n >> 8) % 250}.${n % 250}.${1 + ((n * 7) % 250)}`;
const BASE = `${(__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "")}/api/v1`;
const KEYS = (__ENV.API_KEYS || "").split(",").filter(Boolean);
const DURATION = __ENV.DURATION || "60s";
const TYPES = ["games", "movies", "series", "books"];

const arrival = (rate, exec) => ({
	executor: "constant-arrival-rate",
	rate,
	timeUnit: "1s",
	duration: DURATION,
	preAllocatedVUs: 50,
	maxVUs: 200,
	exec,
});

export const options = {
	scenarios: {
		list: arrival(Number(__ENV.LIST_RPS || 25), "list"),
		detail: arrival(Number(__ENV.DETAIL_RPS || 25), "detail"),
	},
	thresholds: {
		"http_req_duration{scenario:list}": ["p(95)<300"],
		"http_req_duration{scenario:detail}": ["p(95)<150"],
		http_req_failed: ["rate<0.001"],
		checks: ["rate>0.999"],
	},
};

const params = (i) => ({
	headers: {
		Authorization: `Bearer ${KEYS[i % KEYS.length]}`,
		"X-Forwarded-For": fakeIP(i),
	},
});

export function setup() {
	if (KEYS.length < 1) fail("set API_KEYS (qa/k6/seed.sh)");
	const ids = {};
	for (const type of TYPES) {
		const res = http.get(`${BASE}/${type}?limit=100`, params(0));
		if (res.status !== 200) fail(`${type} list answered ${res.status}`);
		ids[type] = res.json("data").map((x) => x.id);
		if (ids[type].length === 0) fail(`no ${type} in the catalogue`);
	}
	return ids;
}

export function list() {
	const type = TYPES[scenario.iterationInTest % TYPES.length];
	const res = http.get(`${BASE}/${type}?limit=20`, params(scenario.iterationInTest));
	check(res, { "list: 200": (r) => r.status === 200 });
}

export function detail(ids) {
	const type = TYPES[scenario.iterationInTest % TYPES.length];
	const pool = ids[type];
	const res = http.get(`${BASE}/${type}/${pool[(scenario.iterationInTest * 7) % pool.length]}`, params(scenario.iterationInTest));
	check(res, { "detail: 200": (r) => r.status === 200 });
}
