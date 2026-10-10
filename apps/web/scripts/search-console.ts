// Google Search Console CLI over a service account (no SDK: a signed JWT is exchanged for an OAuth token).
//   bun run seo:gsc sites                      properties the service account can see
//   bun run seo:gsc sitemaps                   submitted sitemaps and their status
//   bun run seo:gsc submit [url]               submit a sitemap (default: <site>/sitemap.xml)
//   bun run seo:gsc queries [days] [dimension] top rows from Search Analytics (dimension: query|page|country|device)
// Env: GSC_KEY_FILE (service account JSON, gitignored, never committed), GSC_SITE (default sc-domain:hexascore.app).
import { createSign } from "node:crypto";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { z } from "zod";

const env = z
	.object({
		GSC_KEY_FILE: z.string().min(1).default("../../hexascore-b351e788e8ef.json"),
		GSC_SITE: z.string().min(1).default("sc-domain:hexascore.app"),
	})
	.parse(process.env);

const keySchema = z.object({ type: z.literal("service_account"), client_email: z.email(), private_key: z.string() });
const tokenSchema = z.object({ access_token: z.string() });
const TIMEOUT = AbortSignal.timeout.bind(AbortSignal, 15_000);
const API = "https://www.googleapis.com/webmasters/v3";

function b64url(value: string | object): string {
	return Buffer.from(typeof value === "string" ? value : JSON.stringify(value)).toString("base64url");
}

async function accessToken(): Promise<string> {
	const key = keySchema.parse(JSON.parse(readFileSync(resolve(env.GSC_KEY_FILE), "utf8")));
	const now = Math.floor(Date.now() / 1000);
	const unsigned = `${b64url({ alg: "RS256", typ: "JWT" })}.${b64url({
		iss: key.client_email,
		scope: "https://www.googleapis.com/auth/webmasters",
		aud: "https://oauth2.googleapis.com/token",
		iat: now,
		exp: now + 3600,
	})}`;
	const signature = createSign("RSA-SHA256").update(unsigned).sign(key.private_key, "base64url");
	const response = await fetch("https://oauth2.googleapis.com/token", {
		method: "POST",
		headers: { "content-type": "application/x-www-form-urlencoded" },
		body: new URLSearchParams({
			grant_type: "urn:ietf:params:oauth:grant-type:jwt-bearer",
			assertion: `${unsigned}.${signature}`,
		}),
		signal: TIMEOUT(),
	});
	if (!response.ok) throw new Error(`token request failed: HTTP ${response.status} ${await response.text()}`);
	return tokenSchema.parse(await response.json()).access_token;
}

async function call(token: string, path: string, init: RequestInit = {}): Promise<unknown> {
	const response = await fetch(`${API}${path}`, {
		...init,
		headers: { authorization: `Bearer ${token}`, "content-type": "application/json" },
		signal: TIMEOUT(),
	});
	const text = await response.text();
	if (!response.ok) throw new Error(`${init.method ?? "GET"} ${path}: HTTP ${response.status} ${text}`);
	return text ? JSON.parse(text) : {};
}

const site = encodeURIComponent(env.GSC_SITE);
const siteUrl = env.GSC_SITE.startsWith("sc-domain:") ? `https://${env.GSC_SITE.slice(10)}` : env.GSC_SITE;
const [command = "sites", ...args] = process.argv.slice(2);

try {
	const token = await accessToken();
	if (command === "sites") {
		console.log(JSON.stringify(await call(token, "/sites"), null, 2));
	} else if (command === "sitemaps") {
		console.log(JSON.stringify(await call(token, `/sites/${site}/sitemaps`), null, 2));
	} else if (command === "submit") {
		const feed = args[0] ?? `${siteUrl}/sitemap.xml`;
		await call(token, `/sites/${site}/sitemaps/${encodeURIComponent(feed)}`, { method: "PUT" });
		console.log(`submitted ${feed}`);
	} else if (command === "queries") {
		const days = z.coerce
			.number()
			.int()
			.min(1)
			.max(480)
			.parse(args[0] ?? 28);
		const dimension = z.enum(["query", "page", "country", "device"]).parse(args[1] ?? "query");
		const end = new Date();
		const start = new Date(end.getTime() - days * 86_400_000);
		const body = {
			startDate: start.toISOString().slice(0, 10),
			endDate: end.toISOString().slice(0, 10),
			dimensions: [dimension],
			rowLimit: 25,
		};
		console.log(
			JSON.stringify(
				await call(token, `/sites/${site}/searchAnalytics/query`, {
					method: "POST",
					body: JSON.stringify(body),
				}),
				null,
				2,
			),
		);
	} else {
		throw new Error(`unknown command "${command}" (sites | sitemaps | submit [url] | queries [days] [dimension])`);
	}
} catch (error) {
	console.error(`[gsc] ${error instanceof Error ? error.message : String(error)}`);
	process.exit(1);
}
