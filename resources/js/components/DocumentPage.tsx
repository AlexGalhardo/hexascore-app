import { Link, usePage } from "@inertiajs/react";
import type { ReactNode } from "react";

import type { SharedProps } from "../types";
import Seo from "./Seo";

export interface DocumentSection {
	id: string;
	heading: string;
	body: ReactNode;
}

interface Props {
	title: string;
	description: string;
	intro: string;
	sections: DocumentSection[];
	contentsLabel: string;
	updatedLabel?: string;
	updatedAt?: string;
	also?: { label: string; href: string; title: string };
}

const eyebrow = "text-xs font-bold tracking-[0.1em] text-muted uppercase";

// The long-form page layout of /terms, /privacy and /how-score-works: article on the left, sticky contents on the right.
export default function DocumentPage({
	title,
	description,
	intro,
	sections,
	contentsLabel,
	updatedLabel,
	updatedAt,
	also,
}: Props) {
	const { locale } = usePage<SharedProps>().props;

	return (
		<div className="grid gap-12 lg:grid-cols-[minmax(0,45rem)_18.75rem] lg:justify-between">
			<Seo title={title} description={description} />
			<article className="min-w-0">
				<h1 className="text-5xl leading-none md:text-6xl">{title}</h1>
				{updatedAt && (
					<p className="mt-4 flex flex-wrap items-baseline gap-2 text-sm">
						<span className={eyebrow}>{updatedLabel}</span>
						<time dateTime={updatedAt} className="font-medium">
							{new Intl.DateTimeFormat(locale, {
								dateStyle: "long",
								timeZone: "UTC",
							}).format(new Date(updatedAt))}
						</time>
					</p>
				)}
				<p className="mt-6 font-display text-2xl leading-snug italic">{intro}</p>
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
						{s.body}
					</section>
				))}
			</article>
			<aside className="max-lg:order-first lg:sticky lg:top-24 lg:self-start">
				<nav aria-labelledby="document-contents" className="border-t-2 border-fg pt-3">
					<p id="document-contents" className={eyebrow}>
						{contentsLabel}
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
				{also && (
					<>
						<p className={`${eyebrow} mt-8`}>{also.label}</p>
						<Link
							href={also.href}
							className="inline-flex min-h-11 items-center font-medium text-primary hover:underline"
						>
							{also.title}
						</Link>
					</>
				)}
			</aside>
		</div>
	);
}
