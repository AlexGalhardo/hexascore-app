import { router } from "@inertiajs/react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

export type ReactionKind = "like" | "helpful";

interface Props {
	target: "review" | "comment" | "list";
	id: string;
	likes: number;
	helpful: number;
	mine?: ReactionKind | "";
	/** False for anonymous readers and on the reader's own post: counts only. */
	canReact: boolean;
}

const KINDS: ReactionKind[] = ["like", "helpful"];

/** Like and helpful counters; one reaction per reader and target, a second click removes it (SDD 14 §12.11). */
export default function ReactionBar({ target, id, likes, helpful, mine, canReact }: Props) {
	const t = useT();
	const l = useLocalized();
	const url = l(`/reactions/${target}/${id}`);
	const options = { preserveScroll: true };
	const counts = { like: likes, helpful };

	return (
		<span className="flex flex-wrap items-center gap-3 text-sm">
			{KINDS.map((kind) => {
				const label = t(`reactions.${kind}`, { count: counts[kind] });

				return canReact ? (
					<button
						key={kind}
						type="button"
						aria-pressed={mine === kind}
						className={`min-h-11 underline-offset-2 hover:underline ${mine === kind ? "font-semibold text-primary" : "text-muted"}`}
						onClick={() =>
							mine === kind
								? router.delete(url, options)
								: router.put(url, { kind }, options)
						}
					>
						{label}
					</button>
				) : (
					<span key={kind} className="text-muted">
						{label}
					</span>
				);
			})}
		</span>
	);
}
