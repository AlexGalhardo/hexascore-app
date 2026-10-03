import { type NextRequest, NextResponse } from "next/server";
import createMiddleware from "next-intl/middleware";
import { routing } from "./i18n/routing";
import { APP_URL } from "./lib/config";
import { problem } from "./lib/problem";
import { API_HEADERS, createNonce, pageCsp } from "./lib/security-headers";

const intl = createMiddleware(routing);

// The api. host only serves the REST API: its root describes it, anything outside /v1 is a JSON 404, never a page.
function apiHostResponse(pathname: string): Response {
	const headers = Object.fromEntries(API_HEADERS.map((h) => [h.key, h.value]));
	if (pathname === "/") {
		return Response.json(
			{
				name: "HexaScore API",
				version: "v1",
				base_url: `https://api.${new URL(APP_URL).host}/v1`,
				docs: `${APP_URL}/docs`,
				openapi: `${APP_URL}/openapi.json`,
			},
			{ headers },
		);
	}
	return problem(404, "Not Found", "The API lives under /v1. See the docs for every endpoint.", headers);
}

export default function proxy(request: NextRequest): NextResponse | Response {
	const host = request.headers.get("host") ?? "";
	const { pathname } = request.nextUrl;
	// `api.<domain>/v1/...` serves the public API from `/api/v1/...`. It must happen here, not in next.config
	// rewrites: the proxy runs first and next-intl would otherwise prefix `/v1` with a locale.
	if (host.startsWith("api.") && (pathname === "/v1" || pathname.startsWith("/v1/"))) {
		const url = request.nextUrl.clone();
		url.pathname = `/api${pathname}`;
		return NextResponse.rewrite(url);
	}
	if (host.startsWith("api.")) return apiHostResponse(pathname);

	// Next reads the nonce from the request CSP header while rendering and stamps it on its own scripts; next-intl
	// copies `request.headers` into the request it forwards, so setting them here reaches the render.
	const nonce = createNonce();
	const csp = pageCsp(nonce, pathname, process.env.NODE_ENV === "development", process.env.VERCEL_ENV === "preview");
	request.headers.set("x-nonce", nonce);
	request.headers.set("Content-Security-Policy", csp);
	const response = intl(request);
	response.headers.set("Content-Security-Policy", csp);
	return response;
}

export const config = {
	// Skip api, Next internals, files (paths with a literal dot; `.*..*` would skip every path) and the BotID
	// challenge prefix (proxied to Vercel, which sends its own `frame-ancestors 'self'`).
	matcher: ["/((?!api|_next|_vercel|149e9513-01fa-4fb0-aad4-566afd725d1b|.*\\..*).*)"],
};
