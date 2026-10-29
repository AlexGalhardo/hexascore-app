// k6: /search at 20 rps and /search/suggest at 50 rps, 60 s each (SDD 07 NFR-07-1, NFR-07-2, open issue 07-8).
// Start the app with SEARCH_CACHE=off so the anonymous result cache is skipped and every request hits Postgres.
// Env: BASE_URL (default http://localhost:3000). Run:
// docker run --rm --network host -e BASE_URL -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/search.js
import { check } from "k6";
import http from "k6/http";

const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");

const WORDS = [
	"the", "star", "love", "war", "zelda", "mario", "dark", "night", "dragon", "king", "lost", "last", "world",
	"super", "black", "white", "red", "blue", "game", "man", "woman", "city", "dream", "ghost", "fire", "ice",
	"cien", "anos", "senhor", "aneis", "casmurro", "machado", "harry", "potter", "gojira", "acao", "ordem",
	"mundo", "tempo", "vida", "sol", "lua", "mar", "terra", "rei", "guerra", "amor", "noite", "estrela", "pokemon",
];
// 200 distinct search queries: single words, plus every word paired with the next.
const SEARCHES = [...WORDS, ...WORDS.map((w, i) => `${w} ${WORDS[(i + 7) % WORDS.length]}`)];
SEARCHES.push(...WORDS.map((w, i) => `${w} ${WORDS[(i + 13) % WORDS.length]}`));
SEARCHES.push(...WORDS.map((w, i) => `${w} ${WORDS[(i + 21) % WORDS.length]}`));
// All 2-character prefixes of the word list (the worst case for prefix search).
const PREFIXES = [...new Set(WORDS.map((w) => w.slice(0, 2)))];

export const options = {
	scenarios: {
		search: { executor: "constant-arrival-rate", rate: 20, timeUnit: "1s", duration: "60s", preAllocatedVUs: 20, exec: "search" },
		suggest: { executor: "constant-arrival-rate", rate: 50, timeUnit: "1s", duration: "60s", preAllocatedVUs: 20, exec: "suggest" },
	},
	thresholds: {
		"http_req_duration{scenario:search}": ["p(95)<300"],
		"http_req_duration{scenario:suggest}": ["p(95)<50"],
		checks: ["rate>0.99"],
	},
};

export function search() {
	const q = SEARCHES[__ITER % SEARCHES.length];
	const res = http.get(`${BASE}/search?q=${encodeURIComponent(q)}`);
	check(res, { "search: 200": (r) => r.status === 200 });
}

export function suggest() {
	const q = PREFIXES[__ITER % PREFIXES.length];
	const res = http.get(`${BASE}/search/suggest?q=${encodeURIComponent(q)}`);
	check(res, { "suggest: 200": (r) => r.status === 200 });
}
