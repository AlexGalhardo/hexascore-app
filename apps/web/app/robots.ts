import type { MetadataRoute } from "next";
import { APP_URL } from "@/lib/config";
import { disallowedPaths } from "@/lib/seo";

export const dynamic = "force-dynamic";

export default function robots(): MetadataRoute.Robots {
	// Only production is indexable; previews and local builds must never compete with it in search results.
	const indexable = process.env.VERCEL_ENV === "production";
	return {
		rules: indexable
			? { userAgent: "*", allow: "/", disallow: disallowedPaths() }
			: { userAgent: "*", disallow: "/" },
		sitemap: `${APP_URL}/sitemap-index.xml`,
	};
}
