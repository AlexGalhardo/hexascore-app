import { withBotId } from "botid/next/config";
import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";
// Fails the build with an explicit list when env vars are missing or malformed.
import "./lib/env";
import { API_HEADERS, BASE_HEADERS } from "./lib/security-headers";

const withNextIntl = createNextIntlPlugin("./i18n/request.ts");

const nextConfig: NextConfig = {
	reactStrictMode: true,
	poweredByHeader: false,
	typedRoutes: true,
	// PGlite loads its WASM and extension tarballs from its package folder; bundling breaks those paths.
	serverExternalPackages: ["@electric-sql/pglite"],
	// Page CSP (with nonce) is set per request in proxy.ts; BotID adds its own framing headers for its challenge path.
	async headers() {
		return [
			{ source: "/:path*", headers: BASE_HEADERS },
			{ source: "/api/:path*", headers: API_HEADERS },
		];
	},
	// The sitemap index is public at /sitemap.xml only. `app/sitemap.ts` with `generateSitemaps` owns that route id
	// without serving anything at it (a route file there is a "duplicate page" and every metadata route answers 500),
	// so the handler stays at app/sitemap-index.xml: the public URL is rewritten to it and the old URL redirects.
	async redirects() {
		return [{ source: "/sitemap-index.xml", destination: "/sitemap.xml", permanent: true }];
	},
	async rewrites() {
		return [{ source: "/sitemap.xml", destination: "/sitemap-index.xml" }];
	},
};

export default withBotId(withNextIntl(nextConfig));
