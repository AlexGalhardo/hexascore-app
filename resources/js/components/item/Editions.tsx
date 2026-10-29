import { Link, usePage } from "@inertiajs/react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import type { EditionRow } from "./types";

export default function Editions({
	slug,
	editions,
	total,
}: {
	slug: string;
	editions: EditionRow[];
	total: number;
}) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	if (editions.length === 0) return null;

	return (
		<section aria-labelledby="editions-heading" className="grid gap-2">
			<h2 id="editions-heading" className="text-xl font-bold">
				{t("item.editions.title")}
			</h2>
			<ul className="grid gap-1">
				{editions.map((e) => (
					<li key={e.id}>
						<Link href={localized(`/books/${slug}/editions/${e.id}`, locale)}>
							{e.title}
						</Link>
						<span className="ml-2 text-sm text-muted">
							{[e.publisher, e.publicationYear, e.language]
								.filter(Boolean)
								.join(" · ")}
						</span>
					</li>
				))}
			</ul>
			{total > editions.length && (
				<p>{t("item.moreCount", { count: total - editions.length })}</p>
			)}
		</section>
	);
}
