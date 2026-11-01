// k6: moderation queue and suggestion submit with 10,000 open suggestions (ROADMAP P4.B.5.1, SDD 04 NFR-04-1).
// Targets: queue p(95) < 500 ms, submit p(95) < 300 ms.
// LOCAL or staging only. Seed the queue first (the seeder refuses production):
//   ./app db:seed DemoCatalog && ./app db:seed CommunityLoad --count 10000
// Env: BASE_URL (http://localhost:3000), ADMIN_EMAIL (admin@example.test), MEMBERS (comma-separated emails of
// DemoCatalog users, one per submit VU; psql -At -c "select string_agg(email, ',') from (select email from users
// where role = 'user' order by id limit 10) u"), PASSWORD (Demo-password-1), QUEUE_VUS (10), SUBMITS (40; the
// daily quota and the draft burst limit cap each member, so keep SUBMITS <= 4 x MEMBERS), DURATION (60s).
// Needs FLAG_COMMUNITY_SUBMISSIONS=true on the app. /admin needs TOTP (SDD 13 §7.5): without ADMIN_TOTP_SECRET the
// script enrols the admin and logs the secret; pass it as ADMIN_TOTP_SECRET on the next runs.
// The login limit is 5 per 10 minutes per IP; X-Forwarded-For only counts where the app trusts the proxy (compose:
// 172.16.0.0/12; Docker Desktop on Windows reaches the app from 192.168.x, so there keep MEMBERS <= 4).
// docker run --rm --network host -e BASE_URL -e MEMBERS -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/moderation.js
import { check, fail } from "k6";
import { hmac } from "k6/crypto";
import http from "k6/http";
import { Trend } from "k6/metrics";

const BASE = (__ENV.BASE_URL || "http://localhost:3000").replace(/\/$/, "");
const PASSWORD = __ENV.PASSWORD || "Demo-password-1";
const ADMIN = __ENV.ADMIN_EMAIL || "admin@example.test";
const MEMBERS = (__ENV.MEMBERS || "").split(",").filter(Boolean);
const queue = new Trend("moderation_queue", true);
const submit = new Trend("suggestion_submit", true);

export const options = {
	// A VU keeps the session it set up; the default empties the jar every iteration.
	noCookiesReset: true,
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

const json = { "Content-Type": "application/json", Accept: "application/json", ...sameOrigin };

// RFC 6238 TOTP (SHA-1, 30 s, 6 digits) from a base32 secret.
function totp(secret, step = Math.floor(Date.now() / 30000)) {
	let bits = "";
	for (const ch of secret.replace(/[\s=]/g, "").toUpperCase()) {
		bits += "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567".indexOf(ch).toString(2).padStart(5, "0");
	}
	const key = new Uint8Array(Math.floor(bits.length / 8)).map((_, i) => Number.parseInt(bits.slice(i * 8, i * 8 + 8), 2));
	const counter = new Uint8Array(8);
	for (let i = 7, n = step; i >= 0; i--, n = Math.floor(n / 256)) counter[i] = n % 256;
	const d = new Uint8Array(hmac("sha1", key.buffer, counter.buffer, "binary"));
	const o = d[19] & 0x0f;
	const bin = ((d[o] & 0x7f) << 24) | (d[o + 1] << 16) | (d[o + 2] << 8) | d[o + 3];
	return String(bin % 1000000).padStart(6, "0");
}

// The page JSON is in the HTML, quoted or not; Inertia JSON visits would need the asset version.
const prop = (body, re) => re.exec(body.replace(/&quot;/g, '"'))?.[1];

// Every account signs in once in setup() from its own X-Forwarded-For; each VU then reuses the session cookies.
function login(email, i, totpSecret) {
	const jar = http.cookieJar();
	jar.clear(BASE);
	const res = http.post(`${BASE}/login`, JSON.stringify({ identifier: email, password: PASSWORD }), {
		headers: { ...json, "X-Forwarded-For": `10.78.0.${i + 1}` },
		redirects: 0,
		tags: { name: "login" },
	});
	if (res.status >= 400) fail(`login ${email}: ${res.status}`);
	if ((res.headers.Location || "").includes("two-factor-challenge")) {
		if (!totpSecret) fail(`${email} has two-factor on: set ADMIN_TOTP_SECRET`);
		const code = http.post(`${BASE}/two-factor-challenge`, JSON.stringify({ code: totp(totpSecret) }), {
			headers: json,
			redirects: 0,
		});
		if (code.status >= 400) fail(`two-factor ${email}: ${code.status}`);
	}
	const all = jar.cookiesForURL(BASE);
	jar.clear(BASE);
	return Object.fromEntries(Object.keys(all).map((name) => [name, all[name][0]]));
}

// enrolTotp turns TOTP on for the signed-in account and returns the secret.
function enrolTotp(cookies) {
	const jar = http.cookieJar();
	for (const [name, value] of Object.entries(cookies)) jar.set(BASE, name, value);
	http.post(`${BASE}/user/two-factor`, "{}", { headers: json, redirects: 0 });
	const secret = prop(http.get(`${BASE}/user/two-factor`).body, /"secret":"([A-Z2-7=]+)"/);
	if (!secret) fail("no TOTP secret on /user/two-factor");
	const res = http.post(`${BASE}/user/two-factor/confirm`, JSON.stringify({ code: totp(secret) }), {
		headers: json,
		redirects: 0,
	});
	if (res.status >= 400) fail(`two-factor confirm: ${res.status}`);
	console.log(`admin TOTP enrolled; rerun with ADMIN_TOTP_SECRET=${secret}`);
	const all = jar.cookiesForURL(BASE);
	jar.clear(BASE);
	return Object.fromEntries(Object.keys(all).map((name) => [name, all[name][0]]));
}

export function setup() {
	if (MEMBERS.length === 0) fail("MEMBERS is required");
	let admin = login(ADMIN, 0, __ENV.ADMIN_TOTP_SECRET);
	if (!__ENV.ADMIN_TOTP_SECRET) admin = enrolTotp(admin);
	return { admin, members: MEMBERS.map((m, i) => login(m, i + 1)) };
}

let signedIn = false;

function useSession(cookies) {
	if (signedIn) return;
	const jar = http.cookieJar();
	for (const [name, value] of Object.entries(cookies)) jar.set(BASE, name, value);
	signedIn = true;
}

export function queueRead(data) {
	useSession(data.admin);
	const res = http.get(`${BASE}/admin/moderation`, { tags: { name: "queue" } });
	check(res, { "queue 200": (r) => r.status === 200 && r.url.endsWith("/admin/moderation") });
	queue.add(res.timings.duration);
}

export function submitDraft(data) {
	useSession(data.members[(__VU - 1) % data.members.length]);
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
	const mine = http.get(`${BASE}/me/suggestions`, { tags: { name: "mine" } });
	const id = prop(mine.body, /"id":"([0-9a-f-]{36})"/);
	if (!id) return;

	const res = http.post(`${BASE}/suggestions/${id}/submit`, JSON.stringify({ version: 1 }), {
		headers: json,
		redirects: 0,
		tags: { name: "submit" },
	});
	// A quota or cooldown refusal (4xx) is the limiter working, not latency: only a 303 is timed.
	if (res.status === 303) submit.add(res.timings.duration);
}
