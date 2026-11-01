import { Head, usePage } from "@inertiajs/react";

import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

export interface SeoProps {
	title?: string;
	description?: string;
	image?: string;
	type?: string;
	noindex?: boolean;
	// robots overrides the shared value, e.g. "noindex,follow" for thin entity pages.
	robots?: string;
	jsonLd?: object[];
}

export function truncate(text: string, max: number): string {
	return text.length <= max ? text : `${text.slice(0, max - 1).trimEnd()}…`;
}

// "<" is escaped so the data block can never close its own <script>.
export function jsonLdText(data: object[]): string {
	return JSON.stringify(data.length === 1 ? data[0] : data).replace(/</g, String.raw`\u003c`);
}

// Open Graph needs an absolute image URL; a path such as a share card is resolved against the canonical origin.
export function absoluteUrl(src: string, canonical: string | undefined): string {
	return canonical ? new URL(src, canonical).href : src;
}

// Seo sets the document head from the shared `seo` prop (canonical, hreflang,
// robots) and the page's own title, description, image and JSON-LD (SDD 09 §7.7).
export default function Seo({
	title,
	description,
	image,
	type,
	noindex,
	robots: robotsOverride,
	jsonLd,
}: SeoProps) {
	const t = useT();
	const { appName, seo } = usePage<SharedProps>().props;

	const pageTitle = truncate(title ?? (seo?.title || appName), 60);
	const text = truncate(description ?? (seo?.description || t("meta.description")), 160);
	const coverSrc = image ?? seo?.og.image;
	const cover = coverSrc && absoluteUrl(coverSrc, seo?.canonical);
	const robots = robotsOverride ?? (noindex ? "noindex" : seo?.robots);
	const blocks = jsonLd ?? seo?.jsonLd;

	return (
		<Head>
			<title>{pageTitle}</title>
			<meta head-key="description" name="description" content={text} />
			{robots && <meta head-key="robots" name="robots" content={robots} />}
			{seo && <link head-key="canonical" rel="canonical" href={seo.canonical} />}
			{seo &&
				// Lower-case attribute: Inertia's head serializer prints props verbatim.
				Object.entries(seo.alternates).map(([lang, href]) => (
					<link
						key={lang}
						head-key={`alt-${lang}`}
						rel="alternate"
						{...{ hreflang: lang }}
						href={href}
					/>
				))}
			<meta head-key="og:title" property="og:title" content={pageTitle} />
			<meta head-key="og:description" property="og:description" content={text} />
			<meta
				head-key="og:type"
				property="og:type"
				content={type ?? seo?.og.type ?? "website"}
			/>
			{seo && <meta head-key="og:url" property="og:url" content={seo.canonical} />}
			{cover && <meta head-key="og:image" property="og:image" content={cover} />}
			<meta
				head-key="twitter:card"
				name="twitter:card"
				content={cover ? "summary_large_image" : "summary"}
			/>
			{blocks && blocks.length > 0 && (
				<script
					head-key="json-ld"
					type="application/ld+json"
					// biome-ignore lint/security/noDangerouslySetInnerHtml: server-built data, "<" escaped above
					dangerouslySetInnerHTML={{ __html: jsonLdText(blocks) }}
				/>
			)}
		</Head>
	);
}
