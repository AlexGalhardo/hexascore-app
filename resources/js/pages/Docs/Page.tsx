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

const eyebrow = "text-xs font-bold tracking-[0.1em] uppercase";
const navLink =
	"flex min-h-11 items-center border-l-2 px-3 hover:bg-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";
const pager =
	"inline-flex min-h-11 items-center gap-2 font-semibold text-primary hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";

// The Markdown body is styled from here: h1 + first paragraph form the heading and lede, every h2 opens a ruled section.
const prose =
	"min-w-0 max-w-3xl [&_a]:text-primary [&_a]:underline-offset-2 [&_a:hover]:underline [&_blockquote]:my-5 [&_blockquote]:border-l-4 [&_blockquote]:border-primary [&_blockquote]:bg-surface [&_blockquote]:px-4 [&_blockquote]:py-3 [&_code]:font-mono [&_code]:text-[0.9em] [&_h1]:text-5xl [&_h1]:leading-none md:[&_h1]:text-6xl [&>h1+p]:mt-5 [&>h1+p]:text-xl [&>h1+p]:text-muted [&_h2]:mt-8 [&_h2]:border-t-2 [&_h2]:border-fg [&_h2]:pt-4 [&_h2]:text-3xl [&_h3]:mt-6 [&_h3]:font-semibold [&_li]:my-3 [&_li]:ml-5 [&_li]:list-[square] [&_li]:marker:text-fg [&_p]:my-4 [&_p]:leading-relaxed [&_pre]:my-5 [&_pre]:overflow-x-auto [&_pre]:bg-fg [&_pre]:p-5 [&_pre]:text-sm [&_pre]:text-bg [&_table]:my-5 [&_table]:w-full [&_table]:border-collapse [&_table]:text-sm [&_td]:border-b [&_td]:border-border [&_td]:py-2 [&_td]:pr-4 [&_th]:border-b-2 [&_th]:border-fg [&_th]:py-2 [&_th]:pr-4 [&_th]:text-left";

export default function Page({ slug, body, toc, pages }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const path = (s: string) => localized(s === "index" ? "/docs" : `/docs/${s}`, locale);
	const at = pages.indexOf(slug);
	const prev = at > 0 ? pages[at - 1] : null;
	const next = at >= 0 && at < pages.length - 1 ? pages[at + 1] : null;
	const sections = toc.filter((e) => e.level === 2);

	return (
		<>
			<Seo title={`${t(`docs.nav.${slug}`)} - ${t("docs.title")}`} />
			<div className="grid gap-10 md:grid-cols-[15rem_1fr] md:gap-16">
				<nav aria-label={t("docs.navLabel")} className="text-[15px]">
					<p className={`${eyebrow} mb-3 text-muted`}>{t("docs.navLabel")}</p>
					<ul>
						{pages.map((s) => (
							<li key={s}>
								<Link
									href={path(s)}
									aria-current={s === slug ? "page" : undefined}
									className={`${navLink} ${s === slug ? "border-fg bg-surface font-semibold" : "border-transparent"}`}
								>
									{t(`docs.nav.${s}`)}
								</Link>
							</li>
						))}
						<li>
							{/* Scalar is a standalone page outside Inertia, so this is a full navigation. */}
							<a
								href="/docs/reference"
								className={`${navLink} border-transparent font-semibold text-primary`}
							>
								{t("docs.nav.reference")} <span aria-hidden="true">›</span>
							</a>
						</li>
					</ul>
					{sections.length > 0 && (
						<>
							<p className={`${eyebrow} mt-8 mb-2 text-muted`}>
								{t("docs.tocLabel")}
							</p>
							<ul className="text-sm text-muted">
								{sections.map((e) => (
									<li key={e.id}>
										<a
											href={`#${e.id}`}
											className={`${navLink} border-transparent hover:text-fg`}
										>
											{e.text}
										</a>
									</li>
								))}
							</ul>
						</>
					)}
				</nav>
				<div className="min-w-0">
					<p className={`${eyebrow} mb-4 text-muted`}>{t("docs.title")}</p>
					<article
						className={prose}
						// biome-ignore lint/security/noDangerouslySetInnerHtml: build-time escaped Markdown, see Props.body
						dangerouslySetInnerHTML={{ __html: body }}
					/>
					{(prev || next) && (
						<div className="mt-10 flex max-w-3xl justify-between gap-4 border-t border-border pt-6">
							{prev ? (
								<Link href={path(prev)} rel="prev" className={pager}>
									<span aria-hidden="true">‹</span> {t(`docs.nav.${prev}`)}
								</Link>
							) : (
								<span />
							)}
							{next && (
								<Link href={path(next)} rel="next" className={pager}>
									{t(`docs.nav.${next}`)} <span aria-hidden="true">›</span>
								</Link>
							)}
						</div>
					)}
				</div>
			</div>
		</>
	);
}
