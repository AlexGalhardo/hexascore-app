import { Link, usePage } from "@inertiajs/react";

import type { EditionRow } from "../../components/item/types";
import Seo from "../../components/Seo";
import { btnClass } from "../../components/ui/Btn";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	work: { id: string; slug: string; title: string };
	edition: EditionRow;
}

const pencil = (
	<svg
		viewBox="0 0 24 24"
		aria-hidden="true"
		className="size-4 fill-none stroke-current stroke-2"
	>
		<path d="M4 20h4L19 9l-4-4L4 16z" />
	</svg>
);

export default function Edition({ work, edition }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const workHref = localized(`/books/${work.slug}`, locale);
	const rows: [string, string | number | null][] = [
		[t("item.fields.pages"), edition.pageCount],
		[t("item.editionFields.isbn"), edition.isbn13],
		[t("item.editionFields.language"), edition.language],
		[t("item.editionFields.publisher"), edition.publisher],
		[t("item.editionFields.year"), edition.publicationYear],
		[t("item.editionFields.format"), edition.format],
	];

	return (
		<>
			<Seo title={`${edition.title} - ${work.title}`} />
			<nav aria-label={t("item.breadcrumb")} className="mb-10 text-sm text-muted">
				<ol className="flex flex-wrap gap-2">
					<li>
						<Link
							href={localized("/books", locale)}
							className="hover:text-fg hover:underline"
						>
							{t("nav.books")}
						</Link>
					</li>
					<li aria-hidden="true">/</li>
					<li>
						<Link href={workHref} className="hover:text-fg hover:underline">
							{work.title}
						</Link>
					</li>
					<li aria-hidden="true">/</li>
					<li aria-current="page" className="text-fg">
						{t("item.editions.edition")}
					</li>
				</ol>
			</nav>
			<article className="grid max-w-3xl gap-4">
				<p className="text-xs font-bold tracking-[0.1em] text-accent-book uppercase">
					{t("item.editions.edition")} · {t("itemTypes.book")}
				</p>
				<h1 className="text-5xl leading-none md:text-7xl">{edition.title}</h1>
				{edition.subtitle && <p className="text-xl text-muted">{edition.subtitle}</p>}
				<Link
					href={localized(`/suggest/edition/${edition.id}`, locale)}
					className={btnClass("ghost", "-ml-5 justify-self-start")}
				>
					{pencil}
					{t("community.suggest.link")}
				</Link>
				<dl className="mt-4 border-t-2 border-fg">
					{rows
						.filter(([, v]) => v !== null)
						.map(([k, v]) => (
							<div
								key={k}
								className="grid grid-cols-[7rem_minmax(0,1fr)] gap-4 border-b border-border py-2.5 md:grid-cols-[11rem_minmax(0,1fr)]"
							>
								<dt className="text-muted">{k}</dt>
								<dd className="tabular-nums">{v}</dd>
							</div>
						))}
				</dl>
				<Link href={workHref} className={btnClass("ghost", "mt-6 -ml-5 justify-self-start")}>
					<span aria-hidden="true">‹</span>
					{t("item.editions.backToWork")}
				</Link>
			</article>
		</>
	);
}
