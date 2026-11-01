import { Link, usePage } from "@inertiajs/react";

import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	slug: "terms" | "policy";
	updatedAt: string;
}

const eyebrow = "text-xs font-bold tracking-[0.1em] text-muted uppercase";

// Sections are numbered catalogue keys (`<slug>.sections.N.heading`); a missing
// key echoes itself, which ends the list.
export default function Page({ slug, updatedAt }: Props) {
	const t = useT();
	const { appName, locale } = usePage<SharedProps>().props;
	const text = (key: string) => t(key, { appName });
	const sections: { id: string; heading: string; paragraphs: string[] }[] = [];
	// Each page ships only its own namespace, so the link to the other document is labelled from this one.
	const other = slug === "terms" ? "/policy" : "/terms";

	for (let i = 0; ; i++) {
		const base = `${slug}.sections.${i}`;
		if (t(`${base}.heading`) === `${base}.heading`) break;
		const paragraphs: string[] = [];
		for (let j = 0; t(`${base}.body.${j}`) !== `${base}.body.${j}`; j++)
			paragraphs.push(text(`${base}.body.${j}`));
		sections.push({ id: `section-${i + 1}`, heading: text(`${base}.heading`), paragraphs });
	}

	return (
		<div className="grid gap-12 lg:grid-cols-[minmax(0,45rem)_18.75rem] lg:justify-between">
			<Seo title={t(`${slug}.title`)} description={t(`${slug}.description`)} />
			<article className="min-w-0">
				<h1 className="text-5xl leading-none md:text-6xl">{t(`${slug}.title`)}</h1>
				<p className="mt-4 flex flex-wrap items-baseline gap-2 text-sm">
					<span className={eyebrow}>{t("updatedLabel")}</span>
					<time dateTime={updatedAt} className="font-medium">
						{new Intl.DateTimeFormat(locale, {
							dateStyle: "long",
							timeZone: "UTC",
						}).format(new Date(updatedAt))}
					</time>
				</p>
				<p className="mt-6 font-display text-2xl leading-snug italic">
					{text(`${slug}.intro`)}
				</p>
				{sections.map((s) => (
					<section
						key={s.id}
						id={s.id}
						aria-labelledby={`${s.id}-h`}
						className="mt-8 scroll-mt-24 border-t-2 border-fg pt-4"
					>
						<h2 id={`${s.id}-h`} className="mb-3 text-3xl">
							{s.heading}
						</h2>
						{s.paragraphs.map((p) => (
							<p key={p} className="my-4 leading-relaxed">
								{p}
							</p>
						))}
					</section>
				))}
			</article>
			<aside className="max-lg:order-first lg:sticky lg:top-24 lg:self-start">
				<nav aria-labelledby="legal-contents" className="border-t-2 border-fg pt-3">
					<p id="legal-contents" className={eyebrow}>
						{t(`${slug}.contents`)}
					</p>
					<ol className="mt-3 text-[15px] text-muted">
						{sections.map((s) => (
							<li key={s.id}>
								<a
									href={`#${s.id}`}
									className="flex min-h-9 items-center hover:text-fg hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg"
								>
									{s.heading}
								</a>
							</li>
						))}
					</ol>
				</nav>
				<p className={`${eyebrow} mt-8`}>{t(`${slug}.also`)}</p>
				<Link
					href={localized(other, locale)}
					className="inline-flex min-h-11 items-center font-medium text-primary hover:underline"
				>
					{t(`${slug}.otherTitle`)}
				</Link>
			</aside>
		</div>
	);
}
