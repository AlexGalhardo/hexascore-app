import { usePage } from "@inertiajs/react";

import DocumentPage, { type DocumentSection } from "../../components/DocumentPage";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	slug: "terms" | "policy";
	updatedAt: string;
}

// Sections are numbered catalogue keys (`<slug>.sections.N.heading`); a missing
// key echoes itself, which ends the list.
export default function Page({ slug, updatedAt }: Props) {
	const t = useT();
	const { appName, locale } = usePage<SharedProps>().props;
	const text = (key: string) => t(key, { appName });
	const sections: DocumentSection[] = [];
	// Each page ships only its own namespace, so the link to the other document is labelled from this one.
	const other = slug === "terms" ? "/privacy" : "/terms";

	for (let i = 0; ; i++) {
		const base = `${slug}.sections.${i}`;
		if (t(`${base}.heading`) === `${base}.heading`) break;
		const paragraphs: string[] = [];
		for (let j = 0; t(`${base}.body.${j}`) !== `${base}.body.${j}`; j++)
			paragraphs.push(text(`${base}.body.${j}`));
		sections.push({
			id: `section-${i + 1}`,
			heading: text(`${base}.heading`),
			body: paragraphs.map((p) => (
				<p key={p} className="my-4 leading-relaxed">
					{p}
				</p>
			)),
		});
	}

	return (
		<DocumentPage
			title={t(`${slug}.title`)}
			description={t(`${slug}.description`)}
			intro={text(`${slug}.intro`)}
			sections={sections}
			contentsLabel={t(`${slug}.contents`)}
			updatedLabel={t("updatedLabel")}
			updatedAt={updatedAt}
			also={{
				label: t(`${slug}.also`),
				href: localized(other, locale),
				title: t(`${slug}.otherTitle`),
			}}
		/>
	);
}
