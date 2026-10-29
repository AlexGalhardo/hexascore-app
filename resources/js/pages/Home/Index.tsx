import SectionHighlights, { type Section } from "../../components/catalog/SectionHighlights";
import Seo from "../../components/Seo";
import { useT } from "../../lib/useT";

export default function Index({ sections }: { sections: Section[] }) {
	const t = useT();

	return (
		<>
			<Seo />
			<div className="grid gap-12">
				{/* No visible hero by design; the page still needs one h1 for screen readers and search engines. */}
				<h1 className="sr-only">{t("home.title")}</h1>
				{sections.map((section) => (
					<SectionHighlights key={section.type} section={section} />
				))}
			</div>
		</>
	);
}
