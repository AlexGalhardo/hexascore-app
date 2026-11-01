// k6: moderation queue and suggestion submit with 10,000 open suggestions (ROADMAP P4.B.5.1, SDD 04 NFR-04-1).
// Targets: queue p(95) < 500 ms, submit p(95) < 300 ms.
// LOCAL or staging only. Seed the queue first (the seeder refuses production):
//   ./app db:seed DemoCatalog && ./app db:seed CommunityLoad --count 10000
// Env: BASE_URL (http://localhost:3000), ADMIN_EMAIL (admin@example.test), MEMBERS (comma-separated emails of
// DemoCatalog users, one per submit VU; psql -At -c "select string_agg(email, ',') from (select email from users
// where role = 'user' order by id limit 10) u"), PASSWORD (Demo-password-1), QUEUE_VUS (10), SUBMITS (40; the
// daily quota and the draft burst limit cap each member, so keep SUBMITS <= 4 x MEMBERS), DURATION (60s).
// Needs FLAG_COMMUNITY_SUBMISSIONS=true on the app.
// docker run --rm --network host -e BASE_URL -e MEMBERS -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/moderation.js
import { check, fail } from "k6";
import http from "k6/http";
import { Trend } from "k6/metrics";

const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");
const PASSWORD = __ENV.PASSWORD || "Demo-password-1";
const ADMIN = __ENV.ADMIN_EMAIL || "admin@example.test";
const MEMBERS = (__ENV.MEMBERS || "").split(",").filter(Boolean);
const queue = new Trend("moderation_queue", true);
const submit = new Trend("suggestion_submit", true);

export const options = {
	scenarios: {
		queue: {
			executor: "constant-vus",
			exec: "queueRead",
			vus: Number(__ENV.QUEUE_VUS || 10),
			duration: __ENV.DURATION || "60s",
		},
		submit: {
			executor: "shared-iterations",
			exec: "submitDraft",
			vus: Math.max(1, Math.min(MEMBERS.length, 5)),
			iterations: Number(__ENV.SUBMITS || 40),
			maxDuration: "5m",
		},
	},
	thresholds: {
		moderation_queue: ["p(95)<500"],
		suggestion_submit: ["p(95)<300"],
		"checks{scenario:queue}": ["rate>0.99"],
	},
};

// The CSRF middleware accepts a same-origin request (SDD 13); k6 has no browser to add the headers.
const sameOrigin = { Origin: BASE, "Sec-Fetch-Site": "same-origin" };

function login(email) {
	const res = http.post(`${BASE}/login`, JSON.stringify({ identifier: email, password: PASSWORD }), {
		headers: { "Content-Type": "application/json", Accept: "application/json", ...sameOrigin },
		redirects: 0,
		tags: { name: "login" },
	});
	if (res.status >= 400) fail(`login ${email}: ${res.status}`);
}

export function setup() {
	if (MEMBERS.length === 0) fail("MEMBERS is required");
}

let adminIn = false;

export function queueRead() {
	if (!adminIn) {
		login(ADMIN);
		adminIn = true;
	}
	const res = http.get(`${BASE}/admin/moderation`, { tags: { name: "queue" } });
	check(res, { "queue 200": (r) => r.status === 200 });
	queue.add(res.timings.duration);
}

let memberIn = false;

export function submitDraft() {
	if (!memberIn) {
		login(MEMBERS[(__VU - 1) % MEMBERS.length]);
		memberIn = true;
	}
	const json = { "Content-Type": "application/json", Accept: "application/json", ...sameOrigin };
	const draft = http.post(
		`${BASE}/suggestions`,
		JSON.stringify({
			kind: "create",
			domain: "movie",
			locale: "en",
			sourceUrls: ["https://www.wikidata.org/wiki/Q127367"],
			changes: { title: { after: `k6 film ${__VU}-${__ITER}-${Date.now()}` } },
		}),
		{ headers: json, redirects: 0, tags: { name: "draft" } },
	);
	if (draft.status >= 400) return;

	// The save redirects back to the form, so the new draft id comes from the member's own list.
	const mine = http.get(`${BASE}/me/suggestions`, { headers: { "X-Inertia": "true" }, tags: { name: "mine" } });
	const id = /"id":"([0-9a-f-]{36})"/.exec(mine.body)?.[1];
	if (!id) return;

	const res = http.post(`${BASE}/suggestions/${id}/submit`, JSON.stringify({ version: 1 }), {
		headers: json,
		redirects: 0,
		tags: { name: "submit" },
	});
	// A quota or cooldown refusal (4xx) is the limiter working, not latency: only a 303 is timed.
	if (res.status === 303) submit.add(res.timings.duration);
}
