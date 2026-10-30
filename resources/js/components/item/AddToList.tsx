import { router } from "@inertiajs/react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export interface ListChoice {
	id: string;
	title: string;
	hasItem: boolean;
}

export default function AddToList({ itemId, lists }: { itemId: string; lists: ListChoice[] }) {
	const t = useT();
	const loc = useLocalized();

	return (
		<section aria-labelledby="addlist-heading" className="grid gap-2">
			<h2 id="addlist-heading" className="font-semibold">
				{t("lists.addToList")}
			</h2>
			{lists.length === 0 ? (
				<p>{t("lists.noLists")}</p>
			) : (
				<ul className="grid gap-1">
					{lists.map((l) => (
						<li key={l.id} className="flex items-center justify-between gap-2">
							<span>{l.title}</span>
							<button
								type="button"
								className="min-h-11 rounded border px-3"
								onClick={() => {
									const url = loc(`/lists/${l.id}/items/${itemId}`);
									if (l.hasItem) router.delete(url, { preserveScroll: true });
									else router.put(url, {}, { preserveScroll: true });
								}}
							>
								{l.hasItem ? t("lists.remove") : t("lists.addToList")}
							</button>
						</li>
					))}
				</ul>
			)}
		</section>
	);
}
