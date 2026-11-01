import DocumentPage, { type DocumentSection } from "../../components/DocumentPage";
import { useT } from "../../lib/useT";

interface Props {
	categoryKeys: Record<string, string[]>;
}

const topics = [
	"input",
	"formula",
	"minimum",
	"confidence",
	"weights",
	"release",
	"burst",
	"caps",
	"versions",
];

const paragraph = "my-4 leading-relaxed";

// Same layout as /terms (owner round 2026-11-01 item 10); the category list comes from scoring.CategoryKeys.
export default function Score({ categoryKeys }: Props) {
	const t = useT();
	const sections: DocumentSection[] = [
		{
			id: "categories",
			heading: t("help.score.categoriesTitle"),
			body: (
				<>
					<p className={paragraph}>{t("help.score.categoriesIntro")}</p>
					<dl className="space-y-2">
						{Object.entries(categoryKeys).map(([type, keys]) => (
							<div key={type}>
								<dt className="font-medium">{t(`itemTypes.${type}`)}</dt>
								<dd>
									{keys.map((key) => t(`categories.${type}.${key}`)).join(", ")}
								</dd>
							</div>
						))}
					</dl>
				</>
			),
		},
		...topics.map((s) => ({
			id: s,
			heading: t(`help.score.${s}Title`),
			body: <p className={paragraph}>{t(`help.score.${s}`)}</p>,
		})),
	];

	return (
		<DocumentPage
			title={t("help.score.title")}
			description={t("help.score.intro")}
			intro={t("help.score.intro")}
			sections={sections}
			contentsLabel={t("help.score.contents")}
		/>
	);
}
