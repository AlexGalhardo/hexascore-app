// k6: anonymous GET /app/v1/search at 1 rps for 30s. Stays under the 60/min search IP limit.
// Env: BASE_URL (default http://localhost:3000). The k6 binary is not required locally:
// docker run --rm --add-host=host.docker.internal:host-gateway -e BASE_URL=http://host.docker.internal:3000 -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/app-v1.js
import { check } from "k6";
import http from "k6/http";

const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");

export const options = {
	scenarios: {
		search: {
			executor: "constant-arrival-rate",
			rate: 1,
			timeUnit: "1s",
			duration: "30s",
			preAllocatedVUs: 2,
			exec: "search",
		},
	},
	thresholds: {
		http_req_duration: ["p(95)<300"],
		checks: ["rate>0.99"],
	},
};

export function search() {
	const res = http.get(`${BASE}/app/v1/search?q=zelda`);
	check(res, { "search: 200": (r) => r.status === 200 });
}
