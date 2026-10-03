import { type NextRequest, NextResponse } from "next/server";
import createMiddleware from "next-intl/middleware";
import { routing } from "./i18n/routing";

const intl = createMiddleware(routing);

// `api.<domain>/v1/...` serves the public API from `/api/v1/...`. It must happen here, not in next.config
// rewrites: the proxy runs first and next-intl would otherwise prefix `/v1` with a locale.
export default function proxy(request: NextRequest): NextResponse {
	const host = request.headers.get("host") ?? "";
	const { pathname } = request.nextUrl;
	if (host.startsWith("api.") && (pathname === "/v1" || pathname.startsWith("/v1/"))) {
		const url = request.nextUrl.clone();
		url.pathname = `/api${pathname}`;
		return NextResponse.rewrite(url);
	}
	return intl(request);
}

export const config = {
	// Skip api, Next internals and files (paths with a literal dot); `.*..*` would skip every path.
	matcher: ["/((?!api|_next|_vercel|.*\\..*).*)"],
};
