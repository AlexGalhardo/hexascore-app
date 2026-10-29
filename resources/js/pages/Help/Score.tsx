import { Head } from "@inertiajs/react";

import { useT } from "../../lib/useT";

interface Props {
	categoryKeys: Record<string, string[]>;
}

const sections = [
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

export default function Score({ categoryKeys }: Props) {
	const t = useT();

	return (
		<>
			<Head title={t("help.score.metaTitle")} />
			<h1 className="mb-4 text-2xl font-semibold">{t("help.score.title")}</h1>
			<p className="mb-6">{t("help.score.intro")}</p>

			<h2 className="mb-2 text-xl font-semibold">{t("help.score.categoriesTitle")}</h2>
			<p className="mb-2">{t("help.score.categoriesIntro")}</p>
			<dl className="mb-6 space-y-2">
				{Object.entries(categoryKeys).map(([type, keys]) => (
					<div key={type}>
						<dt className="font-medium">{t(`itemTypes.${type}`)}</dt>
						<dd>{keys.map((key) => t(`categories.${type}.${key}`)).join(", ")}</dd>
					</div>
				))}
			</dl>

			{sections.map((s) => (
				<section key={s} className="mb-6">
					<h2 className="mb-2 text-xl font-semibold">{t(`help.score.${s}Title`)}</h2>
					<p>{t(`help.score.${s}`)}</p>
				</section>
			))}
		</>
	);
}
