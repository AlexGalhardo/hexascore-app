import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";
// Fails the build with an explicit list when env vars are missing or malformed.
import "./lib/env";

const withNextIntl = createNextIntlPlugin("./i18n/request.ts");

const nextConfig: NextConfig = {
	reactStrictMode: true,
	poweredByHeader: false,
	typedRoutes: true,
	// PGlite loads its WASM/data files relative to its own module URL; bundling breaks that path.
	serverExternalPackages: ["@electric-sql/pglite"],
};

export default withNextIntl(nextConfig);
