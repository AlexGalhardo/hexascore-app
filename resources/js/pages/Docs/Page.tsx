import { Link, usePage } from "@inertiajs/react";

import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	slug: string;
	title: string;
	// Built by scripts/build-docs.ts: every source character is HTML-escaped
	// there and only a fixed tag set is emitted, so it is safe to inject.
	body: string;
	toc: { id: string; text: string; level: number }[];
	pages: string[];
}

export default function Page({ slug, body, toc, pages }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const path = (s: string) => (s === "index" ? "/docs" : `/docs/${s}`);

	return (
		<>
			<Seo title={`${t(`docs.nav.${slug}`)} - ${t("docs.title")}`} />
			<div className="grid gap-8 md:grid-cols-[14rem_1fr]">
				<nav aria-label={t("docs.navLabel")}>
					<ul className="space-y-1">
						{pages.map((s) => (
							<li key={s}>
								<Link
									href={localized(path(s), locale)}
									aria-current={s === slug ? "page" : undefined}
									className={`inline-flex min-h-11 items-center hover:underline ${s === slug ? "font-semibold" : ""}`}
								>
									{t(`docs.nav.${s}`)}
								</Link>
							</li>
						))}
					</ul>
					{toc.length > 0 && (
						<>
							<p className="mt-4 text-sm font-semibold">{t("docs.tocLabel")}</p>
							<ul className="text-sm">
								{toc
									.filter((e) => e.level === 2)
									.map((e) => (
										<li key={e.id}>
											<a
												href={`#${e.id}`}
												className="inline-flex min-h-11 items-center hover:underline"
											>
												{e.text}
											</a>
										</li>
									))}
							</ul>
						</>
					)}
				</nav>
				<article
					className="max-w-3xl [&_a]:underline [&_code]:font-mono [&_h1]:mb-4 [&_h1]:text-2xl [&_h1]:font-semibold [&_h2]:mt-6 [&_h2]:mb-2 [&_h2]:text-xl [&_h2]:font-semibold [&_h3]:mt-4 [&_h3]:font-semibold [&_li]:ml-5 [&_li]:list-disc [&_p]:my-3 [&_pre]:overflow-x-auto [&_pre]:rounded [&_pre]:border [&_pre]:border-border [&_pre]:p-3 [&_table]:my-3 [&_td]:border [&_td]:border-border [&_td]:p-2 [&_th]:border [&_th]:border-border [&_th]:p-2 [&_th]:text-left"
					// biome-ignore lint/security/noDangerouslySetInnerHtml: build-time escaped Markdown, see Props.body
					dangerouslySetInnerHTML={{ __html: body }}
				/>
			</div>
		</>
	);
}
