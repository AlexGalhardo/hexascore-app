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
};

export default withBotId(withNextIntl(nextConfig));
