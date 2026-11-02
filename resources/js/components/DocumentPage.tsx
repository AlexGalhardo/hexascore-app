import { Link, usePage } from "@inertiajs/react";
import { type ReactNode, useEffect, useState } from "react";

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

// The first section that intersects the reading band (top 20% to 30% of the viewport) is the current one.
// Without IntersectionObserver (old browsers, jsdom) no section is highlighted and the links still work.
function useCurrentSection(ids: string[]): string | undefined {
	const [current, setCurrent] = useState<string>();
	const key = ids.join(" ");

	useEffect(() => {
		if (typeof IntersectionObserver === "undefined") return;
		const order = key.split(" ");
		const visible = new Set<string>();
		const observer = new IntersectionObserver(
			(entries) => {
				for (const entry of entries) {
					if (entry.isIntersecting) visible.add(entry.target.id);
					else visible.delete(entry.target.id);
				}
				setCurrent(order.find((id) => visible.has(id)));
			},
			{ rootMargin: "-20% 0px -70% 0px" },
		);
		for (const id of order) {
			const el = document.getElementById(id);
			if (el) observer.observe(el);
		}
		return () => observer.disconnect();
	}, [key]);

	return current;
}

// The long-form page layout of /terms, /privacy, /how-score-works and /about: sticky contents on the left
// (above the text below the lg breakpoint), article at reading width on the right.
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
	const current = useCurrentSection(sections.map((s) => s.id));

	return (
		<div className="mx-auto grid max-w-[66rem] gap-12 lg:grid-cols-[15rem_minmax(0,45rem)] lg:justify-center lg:gap-16">
			<Seo title={title} description={description} />
			<aside className="lg:sticky lg:top-24 lg:self-start">
				<nav aria-labelledby="document-contents">
					<p id="document-contents" className={eyebrow}>
						{contentsLabel}
					</p>
					<ol className="mt-3 text-[15px]">
						{sections.map((s) => {
							const isCurrent = s.id === current;
							return (
								<li key={s.id}>
									<a
										href={`#${s.id}`}
										aria-current={isCurrent ? "location" : undefined}
										className={`flex min-h-11 items-center hover:text-fg hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg ${isCurrent ? "font-semibold text-fg underline underline-offset-4" : "text-muted"}`}
									>
										{s.heading}
									</a>
								</li>
							);
						})}
					</ol>
				</nav>
			</aside>
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
				<p className="mt-6 font-display text-2xl leading-snug">{intro}</p>
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
				{also && (
					<div className="mt-10 border-t-2 border-fg pt-4">
						<p className={eyebrow}>{also.label}</p>
						<Link
							href={also.href}
							className="inline-flex min-h-11 items-center font-medium text-primary hover:underline"
						>
							{also.title}
						</Link>
					</div>
				)}
			</article>
		</div>
	);
}
