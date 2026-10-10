import type { MetadataRoute } from "next";
import { APP_URL } from "@/lib/config";
import { env } from "@/lib/env";
import { disallowedPaths } from "@/lib/seo";

export const dynamic = "force-dynamic";

export default function robots(): MetadataRoute.Robots {
	// Only production is indexable; previews, the sandbox (fake data, even if deployed as a Vercel production) and
	// local builds must never compete with it in search results.
	const indexable = process.env.VERCEL_ENV === "production" && !env.NEXT_PUBLIC_SANDBOX;
	return {
		rules: indexable
			? { userAgent: "*", allow: "/", disallow: disallowedPaths() }
			: { userAgent: "*", disallow: "/" },
		sitemap: `${APP_URL}/sitemap.xml`,
	};
}
