// k6: GET /feed at 30 rps, p(95) <= 300 ms with 10k events (SDD 14 NFR-14-7). Needs FLAG_SOCIAL=true and the
// DemoCatalog seed plus qa/k6/feed-seed.sql. setup() signs in USERS (DemoCatalog nicknames) once each with a distinct
// X-Forwarded-For, so the login proof-of-work gate does not trip (the app must trust the proxy, see docker-compose.yml).
// Env: BASE_URL, USERS (demoaaab,demoaaac,demoaaad,demoaaae,demoaaaf), PASSWORD (Demo-password-1), RPS (30),
// DURATION (60s). Run:
// docker run --rm --network host -e BASE_URL -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/feed.js
import { check, fail } from "k6";
import { scenario } from "k6/execution";
import http from "k6/http";

const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");
const USERS = (__ENV.USERS || "demoaaab,demoaaac,demoaaad,demoaaae,demoaaaf").split(",").filter(Boolean);
const PASSWORD = __ENV.PASSWORD || "Demo-password-1";

export const options = {
	scenarios: {
		feed: {
			executor: "constant-arrival-rate",
			rate: Number(__ENV.RPS || 30),
			timeUnit: "1s",
			duration: __ENV.DURATION || "60s",
			preAllocatedVUs: 30,
			maxVUs: 120,
		},
	},
	thresholds: {
		"http_req_duration{name:feed}": ["p(95)<300"],
		http_req_failed: ["rate<0.001"],
		checks: ["rate>0.999"],
	},
};

export function setup() {
	const jar = http.cookieJar();
	const cookies = USERS.map((user, i) => {
		jar.clear(BASE);
		const res = http.post(`${BASE}/login`, JSON.stringify({ identifier: user, password: PASSWORD }), {
			headers: {
				"Content-Type": "application/json",
				Accept: "application/json",
				"X-Forwarded-For": `10.77.0.${i + 1}`,
			},
			redirects: 0,
		});
		if (res.status >= 400) fail(`login ${user} answered ${res.status}`);
		const all = jar.cookiesForURL(BASE);
		return Object.keys(all)
			.map((name) => `${name}=${all[name][0]}`)
			.join("; ");
	});
	if (cookies.length === 0) fail("set USERS");
	return cookies;
}

export default function (cookies) {
	const res = http.get(`${BASE}/feed`, {
		headers: { Cookie: cookies[scenario.iterationInTest % cookies.length] },
		redirects: 0,
		tags: { name: "feed" },
	});
	check(res, { "200": (r) => r.status === 200 });
}
