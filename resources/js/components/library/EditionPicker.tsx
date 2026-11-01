import { router } from "@inertiajs/react";
import { useInfiniteQuery } from "@tanstack/react-query";
import { useState } from "react";
import { withQueryClient } from "../../lib/queryClient";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

interface Edition {
	id: string;
	title: string;
	language: string;
	publisher: string | null;
	publicationYear: number | null;
}

const label = (e: Edition): string =>
	[e.title, e.language, e.publisher, e.publicationYear].filter(Boolean).join(" · ");

function EditionPicker({ itemId }: { itemId: string }) {
	const t = useT();
	const l = useLocalized();
	const [started, setStarted] = useState(false);
	const [choice, setChoice] = useState("");
	const query = useInfiniteQuery({
		queryKey: ["editions", itemId],
		queryFn: async ({ pageParam, signal }): Promise<{ data: Edition[]; next: number }> => {
			const res = await fetch(`/items/${itemId}/editions?cursor=${pageParam}`, {
				signal,
				headers: { Accept: "application/json" },
			});
			if (!res.ok) throw new Error(String(res.status));
			return res.json();
		},
		initialPageParam: 0,
		getNextPageParam: (last) => (last.next >= 0 ? last.next : undefined),
		enabled: started,
	});
	const editions = query.data?.pages.flatMap((p) => p.data) ?? null;
	const failed = query.isError;
	const selected = choice || (editions?.[0]?.id ?? "");
	// cancelRefetch false: a second click while a page loads must not fetch it again.
	const start = () => (started ? void query.refetch({ cancelRefetch: false }) : setStarted(true));
	const more = () => void query.fetchNextPage({ cancelRefetch: false });

	if (editions === null) {
		return (
			<span className="text-sm">
				<button type="button" className="underline" onClick={start}>
					{t("library.chooseEdition")}
				</button>
				{failed && (
					<span role="alert" className="ml-2 text-danger">
						{t("library.editionsError")}
					</span>
				)}
			</span>
		);
	}

	if (editions.length === 0) {
		return <span className="text-sm text-muted">{t("library.noEditions")}</span>;
	}

	return (
		<span className="flex items-center gap-2 text-sm">
			<select
				aria-label={t("library.pickEdition")}
				value={selected}
				onChange={(e) => setChoice(e.target.value)}
				className="rounded-md border border-border bg-surface px-2 py-1"
			>
				{editions.map((e) => (
					<option key={e.id} value={e.id}>
						{label(e)}
					</option>
				))}
			</select>
			{query.hasNextPage && (
				<button type="button" className="underline" onClick={more}>
					{t("library.moreEditions")}
				</button>
			)}
			{failed && (
				<span role="alert" className="text-danger">
					{t("library.editionsError")}
				</span>
			)}
			<button
				type="button"
				className="underline"
				onClick={() =>
					router.put(
						l(`/items/${itemId}/edition`),
						{ editionId: selected },
						{ preserveScroll: true },
					)
				}
			>
				{t("library.saveEdition")}
			</button>
		</span>
	);
}

export default withQueryClient(EditionPicker);
