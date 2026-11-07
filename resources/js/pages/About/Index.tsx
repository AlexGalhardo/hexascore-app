import { Link, usePage } from "@inertiajs/react";
import type { ReactElement } from "react";

import CategoryIcon, { type CategoryType } from "../../components/CategoryIcon";
import { SECTION_PATHS } from "../../components/catalog/SectionHighlights";
import DocumentPage, { type DocumentSection } from "../../components/DocumentPage";
import { ACCENT_TEXT } from "../../components/home/TypeSection";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const TYPES: CategoryType[] = ["game", "movie", "series", "book"];
const SOURCE_URL = "https://github.com/AlexGalhardo/hexascore-app";
const textLink =
	"inline-flex min-h-11 items-center font-semibold text-primary hover:underline outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50";

// Copy from README and SDD 09 (owner round 2026-11-01 item 25); same layout as the legal pages.
export default function Index(): ReactElement {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const section = (
		id: string,
		body: ReactElement | null,
		link?: { href: string; label: string; external?: boolean },
	): DocumentSection => ({
		id,
		heading: t(`about.${id}Title`),
		body: (
			<>
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
			</>
		),
	});

	const sections: DocumentSection[] = [
		section(
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
		),
		section("scores", null, {
			href: localized("/how-score-works", locale),
			label: t("about.scoresLink"),
		}),
		section("community", null),
		section("api", null, { href: "/docs", label: t("about.apiLink") }),
		section("openSource", null, {
			href: SOURCE_URL,
			label: t("about.openSourceLink"),
			external: true,
		}),
		section("contact", null, {
			href: localized("/contact", locale),
			label: t("about.contactLink"),
		}),
		// IGDB and TMDB attribution (ADR 0006); TMDB accepts it in an About or Credits section (owner 2026-11-02).
		section("credits", null),
	];

	return (
		<DocumentPage
			title={t("about.title")}
			description={t("about.description")}
			intro={t("about.intro")}
			sections={sections}
			contentsLabel={t("about.contents")}
		/>
	);
}
