import { router } from "@inertiajs/react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import { btnClass } from "../ui/Btn";
import Dropdown from "../ui/Dropdown";

export interface ListChoice {
	id: string;
	title: string;
	hasItem: boolean;
}

// "Add to list" button that opens the viewer's lists; each row toggles membership.
export default function AddToList({
	itemId,
	lists,
	full = false,
}: {
	itemId: string;
	lists: ListChoice[];
	full?: boolean;
}) {
	const t = useT();
	const loc = useLocalized();

	return (
		<Dropdown
			trigger={t("lists.addToList")}
			label={t("lists.addToList")}
			triggerClassName={btnClass("secondary", full ? "w-full" : "")}
			panelClassName="left-0"
		>
			{lists.length === 0 ? (
				<p className="px-4 py-2 text-sm text-muted">{t("lists.noLists")}</p>
			) : (
				lists.map((l) => (
					<button
						key={l.id}
						type="button"
						aria-pressed={l.hasItem}
						className="flex min-h-11 w-full items-center justify-between gap-4 px-4 text-left text-sm hover:bg-hover focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary"
						onClick={() => {
							const url = loc(`/lists/${l.id}/items/${itemId}`);
							if (l.hasItem) router.delete(url, { preserveScroll: true });
							else router.put(url, {}, { preserveScroll: true });
						}}
					>
						<span>{l.title}</span>
						<span className="text-xs text-muted">
							{l.hasItem ? t("lists.remove") : t("lists.addToList")}
						</span>
					</button>
				))
			)}
		</Dropdown>
	);
}
