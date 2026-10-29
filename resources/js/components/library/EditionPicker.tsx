import { router } from "@inertiajs/react";
import { useState } from "react";

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

export default function EditionPicker({ itemId }: { itemId: string }) {
	const t = useT();
	const [editions, setEditions] = useState<Edition[] | null>(null);
	const [failed, setFailed] = useState(false);
	const [choice, setChoice] = useState("");
	const [next, setNext] = useState(-1);

	const load = async (cursor: number) => {
		setFailed(false);
		try {
			const res = await fetch(`/items/${itemId}/editions?cursor=${cursor}`, {
				headers: { Accept: "application/json" },
			});
			if (!res.ok) throw new Error(String(res.status));
			const body = (await res.json()) as { data: Edition[]; next: number };
			setEditions((prev) => [...(prev ?? []), ...body.data]);
			if (cursor === 0) setChoice(body.data[0]?.id ?? "");
			setNext(body.next);
		} catch {
			setFailed(true);
		}
	};

	if (editions === null) {
		return (
			<span className="text-sm">
				<button type="button" className="underline" onClick={() => load(0)}>
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
				value={choice}
				onChange={(e) => setChoice(e.target.value)}
				className="rounded-md border border-border bg-surface px-2 py-1"
			>
				{editions.map((e) => (
					<option key={e.id} value={e.id}>
						{label(e)}
					</option>
				))}
			</select>
			{next >= 0 && (
				<button type="button" className="underline" onClick={() => load(next)}>
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
						`/items/${itemId}/edition`,
						{ editionId: choice },
						{ preserveScroll: true },
					)
				}
			>
				{t("library.saveEdition")}
			</button>
		</span>
	);
}
