// k6 load + quota/rate-limit check for the public API v1 (docs/testing.md → "Carga").
// Env: BASE_URL, QA_API_KEY (Pro/Business), optional QA_FREE_API_KEY, VERCEL_AUTOMATION_BYPASS_SECRET, RATE.
import { check, fail, sleep } from "k6";
import http from "k6/http";
import { Counter } from "k6/metrics";

const BASE = `${(__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "")}/api/v1`;
// Requests per second of the load phase; must stay under the paid key's limit (Pro 10/s).
const RATE = Number(__ENV.RATE || 8);
const FREE_KEY = __ENV.QA_FREE_API_KEY || "";

const rateLimited = new Counter("rate_limited_429");
const quotaExhausted = new Counter("quota_exhausted_402");

function headers(key) {
	const h = { Authorization: `Bearer ${key}` };
	if (__ENV.VERCEL_AUTOMATION_BYPASS_SECRET) h["x-vercel-protection-bypass"] = __ENV.VERCEL_AUTOMATION_BYPASS_SECRET;
	return { headers: h };
}

const scenarios = {
	load: { executor: "constant-arrival-rate", rate: RATE, timeUnit: "1s", duration: "60s", preAllocatedVUs: 20 },
	// After the load phase so its 429s do not pollute the latency numbers.
	burst: { executor: "per-vu-iterations", vus: 1, iterations: 1, startTime: "62s", exec: "burst" },
};
const thresholds = {
	"http_req_duration{scenario:load}": ["p(95)<300"],
	"checks{scenario:load}": ["rate>0.99"],
	"checks{scenario:burst}": ["rate==1"],
	rate_limited_429: ["count>0"],
};
if (FREE_KEY) {
	// Free = 100 requests for life at 1 req/s: the first run spends what is left (≤ ~110 s), later runs hit 402 at once.
	scenarios.quota = { executor: "per-vu-iterations", vus: 1, iterations: 1, maxDuration: "4m", exec: "quota" };
	thresholds["checks{scenario:quota}"] = ["rate==1"];
	thresholds.quota_exhausted_402 = ["count==1"];
}
export const options = { scenarios, thresholds };

export function setup() {
	const params = headers(__ENV.QA_API_KEY);
	const usage = http.get(`${BASE}/me/usage`, params);
	if (usage.status !== 200) fail(`QA_API_KEY rejected: /me/usage answered ${usage.status}`);
	const { plan, rate_limit } = usage.json();
	if (plan === "free") fail("QA_API_KEY is on the Free plan; use a Pro or Business key");
	if (RATE >= rate_limit.requests_per_second)
		fail(`RATE=${RATE} must stay under ${rate_limit.requests_per_second} req/s`);
	const ids = {};
	for (const type of ["games", "movies", "series", "books"]) {
		const list = http.get(`${BASE}/${type}?limit=10`, params);
		if (list.status !== 200) fail(`/${type} answered ${list.status}`);
		ids[type] = list.json("data").map((item) => item.id);
	}
	return { ids, rps: rate_limit.requests_per_second };
}

const TYPES = ["games", "movies", "series", "books"];
const QUERIES = ["the", "star", "love", "war", "zelda"];

export default function (data) {
	const type = TYPES[__ITER % TYPES.length];
	const ids = data.ids[type];
	const paths = [
		`/${type}?limit=10`,
		`/${type}?sort=-score&min_score=60&limit=10`,
		`/${type}/genres`,
		`/search?q=${QUERIES[__ITER % QUERIES.length]}&limit=10`,
		ids.length ? `/${type}/${ids[__ITER % ids.length]}` : `/${type}?limit=5`,
	];
	const res = http.get(`${BASE}${paths[__ITER % paths.length]}`, headers(__ENV.QA_API_KEY));
	check(res, { "load: 200": (r) => r.status === 200 });
}

// More requests in one instant than the plan's per-second limit: some must be throttled with a 429 problem.
export function burst(data) {
	const params = headers(__ENV.QA_API_KEY);
	const responses = http.batch(
		Array.from({ length: data.rps * 3 }, () => ["GET", `${BASE}/games/genres`, null, params]),
	);
	const throttled = responses.filter((r) => r.status === 429);
	rateLimited.add(throttled.length);
	check(responses, {
		"burst: only 200 or 429": (rs) => rs.every((r) => r.status === 200 || r.status === 429),
		"burst: 429 has Retry-After and a problem body": () =>
			throttled.length > 0 &&
			throttled.every((r) => Number(r.headers["Retry-After"]) > 0 && r.json("status") === 429),
	});
	sleep(2);
}

// Free key: spend the lifetime quota at 1 req/s, then the next request must be 402; `limit=11` is a 400.
export function quota() {
	const params = headers(FREE_KEY);
	const tooBig = http.get(`${BASE}/games?limit=11`, params);
	check(tooBig, { "quota: free limit=11 is 400": (r) => r.status === 400 });
	let res;
	for (let i = 0; i < 120; i++) {
		res = http.get(`${BASE}/games/genres`, params);
		if (res.status === 429) {
			sleep(1.1);
			continue;
		}
		if (res.status !== 200) break;
		sleep(1.05);
	}
	if (res.status === 402) quotaExhausted.add(1);
	check(res, {
		"quota: exhausted free key gets 402": (r) => r.status === 402,
		"quota: 402 is a problem with code": (r) => r.json("status") === 402 && typeof r.json("code") === "string",
	});
}
