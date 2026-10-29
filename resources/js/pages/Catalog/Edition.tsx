import { Link, usePage } from "@inertiajs/react";

import type { EditionRow } from "../../components/item/types";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	work: { id: string; slug: string; title: string };
	edition: EditionRow;
}

export default function Edition({ work, edition }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
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
			<article className="grid gap-4">
				<h1 className="text-3xl font-bold">{edition.title}</h1>
				{edition.subtitle && <p className="text-muted">{edition.subtitle}</p>}
				<dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1">
					{rows
						.filter(([, v]) => v !== null)
						.map(([k, v]) => (
							<div key={k} className="contents">
								<dt>{k}</dt>
								<dd>{v}</dd>
							</div>
						))}
				</dl>
				<Link href={localized(`/books/${work.slug}`, locale)}>
					{t("item.editions.backToWork")}
				</Link>
			</article>
		</>
	);
}
