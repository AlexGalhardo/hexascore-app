import { usePage } from "@inertiajs/react";

import Seo from "../../components/Seo";
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
	const sections: { heading: string; paragraphs: string[] }[] = [];

	for (let i = 0; ; i++) {
		const base = `${slug}.sections.${i}`;
		if (t(`${base}.heading`) === `${base}.heading`) break;
		const paragraphs: string[] = [];
		for (let j = 0; t(`${base}.body.${j}`) !== `${base}.body.${j}`; j++)
			paragraphs.push(text(`${base}.body.${j}`));
		sections.push({ heading: text(`${base}.heading`), paragraphs });
	}

	return (
		<article className="mx-auto max-w-3xl">
			<Seo title={t(`${slug}.title`)} description={t(`${slug}.description`)} />
			<h1 className="text-2xl font-semibold">{t(`${slug}.title`)}</h1>
			<p className="mt-1 text-sm text-muted">
				{t("updatedLabel")}:{" "}
				<time dateTime={updatedAt}>
					{new Intl.DateTimeFormat(locale, { dateStyle: "long", timeZone: "UTC" }).format(
						new Date(updatedAt),
					)}
				</time>
			</p>
			<p className="my-4 text-lg">{text(`${slug}.intro`)}</p>
			{sections.map((s) => (
				<section key={s.heading} className="mb-6">
					<h2 className="mb-2 text-xl font-semibold">{s.heading}</h2>
					{s.paragraphs.map((p) => (
						<p key={p} className="my-2 leading-relaxed">
							{p}
						</p>
					))}
				</section>
			))}
		</article>
	);
}
