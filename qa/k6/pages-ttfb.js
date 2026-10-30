// k6: SSR TTFB p75 <= 400 ms on home, listing, item and entity pages (SDD 09 NFR-09-4, open issue 09-3). 10 rps per page kind.
// Env: BASE_URL, API_KEY (any key, only to pick item slugs; seeded by qa/k6/seed.sh), ENTITY_PATHS (default demo-catalogue
// people/platform pages), DURATION (60s). Run:
// docker run --rm --network host -e BASE_URL -e API_KEY -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/pages-ttfb.js
import { check, fail } from "k6";
import { scenario } from "k6/execution";
import http from "k6/http";

// Distinct client address per request so the per-IP limits do not mask latency.
const fakeIP = (n) => `10.${(n >> 8) % 250}.${n % 250}.${1 + ((n * 7) % 250)}`;
const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");
const ENTITIES = (__ENV.ENTITY_PATHS || "/people/ana-oliveira,/platforms/android,/companies/allen-studio-llc")
	.split(",")
	.filter(Boolean);
const KINDS = ["home", "listing", "item", "entity"];

const scenarios = {};
const thresholds = { http_req_failed: ["rate<0.005"], checks: ["rate>0.995"] };
for (const kind of KINDS) {
	scenarios[kind] = {
		executor: "constant-arrival-rate",
		rate: 10,
		timeUnit: "1s",
		duration: __ENV.DURATION || "60s",
		preAllocatedVUs: 10,
		maxVUs: 50,
		exec: kind,
	};
	thresholds[`http_req_waiting{scenario:${kind}}`] = ["p(75)<400"];
}
export const options = { scenarios, thresholds };

export function setup() {
	const key = (__ENV.API_KEY || "").split(",")[0];
	if (!key) fail("set API_KEY");
	const res = http.get(`${BASE}/api/v1/games?limit=100`, { headers: { Authorization: `Bearer ${key}` } });
	if (res.status !== 200) fail(`games list answered ${res.status}`);
	return { slugs: res.json("data").map((x) => x.slug) };
}

const get = (path) =>
	http.get(`${BASE}${path}`, {
		headers: { "X-Forwarded-For": fakeIP(scenario.iterationInTest) },
	});
const ok = (res) => check(res, { "200": (r) => r.status === 200 });

export function home() {
	ok(get("/"));
}
export function listing() {
	ok(get(`/games?page=${1 + (scenario.iterationInTest % 5)}`));
}
export function item(data) {
	ok(get(`/games/${data.slugs[scenario.iterationInTest % data.slugs.length]}`));
}
export function entity() {
	ok(get(ENTITIES[scenario.iterationInTest % ENTITIES.length]));
}
