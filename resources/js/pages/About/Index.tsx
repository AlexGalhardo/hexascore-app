import { Link, usePage } from "@inertiajs/react";
import type { ReactElement } from "react";

import CategoryIcon, { type CategoryType } from "../../components/CategoryIcon";
import { SECTION_PATHS } from "../../components/catalog/SectionHighlights";
import { ACCENT_TEXT } from "../../components/home/TypeSection";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const TYPES: CategoryType[] = ["game", "movie", "series", "book"];
const SOURCE_URL = "https://github.com/AlexGalhardo/hexascore-app";
const textLink =
	"inline-flex min-h-11 items-center font-semibold text-primary hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";

// Copy from README and SDD 09 (owner round 2026-11-01 item 25); reading width of the legal pages (45rem).
export default function Index(): ReactElement {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const section = (
		id: string,
		body: ReactElement | null,
		link?: { href: string; label: string; external?: boolean },
	): ReactElement => (
		<section aria-labelledby={`about-${id}`} className="mt-8 border-t-2 border-fg pt-4">
			<h2 id={`about-${id}`} className="mb-3 text-3xl">
				{t(`about.${id}Title`)}
			</h2>
			<p className="my-4 leading-relaxed">{t(`about.${id}Body`)}</p>
			{body}
			{link &&
				(link.external ? (
					<a
						href={link.href}
						target="_blank"
						rel="noopener noreferrer"
						className={textLink}
					>
						{link.label}
					</a>
				) : (
					<Link href={link.href} className={textLink}>
						{link.label}
					</Link>
				))}
		</section>
	);

	return (
		<article className="max-w-[45rem]">
			<Seo title={t("about.title")} description={t("about.description")} />
			<h1 className="text-5xl leading-none md:text-6xl">{t("about.title")}</h1>
			<p className="mt-6 font-display text-2xl leading-snug italic">{t("about.intro")}</p>
			{section(
				"categories",
				<ul className="grid gap-x-6 sm:grid-cols-2">
					{TYPES.map((type) => (
						<li key={type}>
							<Link
								href={localized(`/${SECTION_PATHS[type]}`, locale)}
								className={`inline-flex min-h-11 items-center gap-2 font-semibold hover:underline ${ACCENT_TEXT[type]}`}
							>
								<CategoryIcon type={type} />
								{t(`catalog.title.${type}`)}
							</Link>
						</li>
					))}
				</ul>,
			)}
			{section("scores", null, {
				href: localized("/how-score-works", locale),
				label: t("about.scoresLink"),
			})}
			{section("community", null)}
			{section("api", null, { href: "/docs", label: t("about.apiLink") })}
			{section("openSource", null, {
				href: SOURCE_URL,
				label: t("about.openSourceLink"),
				external: true,
			})}
			{section("contact", null, {
				href: localized("/contact", locale),
				label: t("about.contactLink"),
			})}
		</article>
	);
}
