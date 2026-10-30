// k6: 50 VUs hammering POST /login with the proof-of-work gate active (SDD 13 §10, ADR 0016). Every VU
// shares the runner's IP, so the challenge trips after 5 attempts and the rest of the run measures the
// 428 path (counter + HMAC check + fresh challenge). Watch the app RSS while it runs (target ≤ 512 MiB).
// Env: BASE_URL. Run: docker run --rm --network host -e BASE_URL -v "$PWD/qa:/qa" grafana/k6:2.3.0 run
// /qa/k6/login-pow.js
import { check } from "k6";
import http from "k6/http";

const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");

export const options = {
	scenarios: {
		login: { executor: "constant-vus", vus: 50, duration: "60s" },
	},
	thresholds: {
		"http_req_duration{name:login}": ["p(95)<200"],
		checks: ["rate>0.99"],
	},
};

export default function () {
	const probe = http.get(`${BASE}/challenge?route=login`, {
		headers: { Accept: "application/json" },
		tags: { name: "challenge" },
	});
	check(probe, { "challenge 200": (r) => r.status === 200 });

	const res = http.post(
		`${BASE}/login`,
		JSON.stringify({ identifier: `k6-${__VU}@example.com`, password: "not-the-password" }),
		{
			headers: { "Content-Type": "application/json", Accept: "application/json" },
			tags: { name: "login" },
		},
	);
	// Below the threshold a wrong password is refused; past it the gate answers 428. Never a 5xx.
	check(res, { "login refused, not crashed": (r) => r.status < 500 });
}
