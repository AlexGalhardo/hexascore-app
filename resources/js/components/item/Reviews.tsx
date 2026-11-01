import { useState } from "react";

import { useT } from "../../lib/useT";
import CommentThread, { type CommentEntry } from "../social/CommentThread";
import Markdown, { type MdNode } from "../social/Markdown";
import ReactionBar, { type ReactionKind } from "../social/ReactionBar";
import ReviewForm from "./ReviewForm";

/** One review card from the deferred `reviews` prop (SDD 14 §12.10); `source` only on the viewer's own. */
export interface ReviewEntry {
	id: string;
	author: { nickname: string; initial: string };
	item?: { id: string; type: string; slug: string; title: string };
	stars: number | null;
	body: MdNode[];
	source?: string;
	spoiler: boolean;
	helpful: number;
	likes?: number;
	/** Item page only (P7.B.1.2). */
	myReaction?: ReactionKind;
	comments?: CommentEntry[];
	createdAt: string;
	updatedAt: string;
}

interface Props {
	itemId: string;
	reviews: { data: ReviewEntry[]; mine: ReviewEntry | null };
	canWrite: boolean;
	rated: boolean;
}

export function ReviewCard({ review }: { review: ReviewEntry }) {
	const t = useT();
	const [shown, setShown] = useState(!review.spoiler);

	return (
		<article className="space-y-2 py-3">
			<header className="flex flex-wrap items-baseline gap-2 text-sm">
				<span className="font-semibold">{review.author.nickname}</span>
				{review.stars !== null && (
					<span className="text-muted">{review.stars.toFixed(1)}</span>
				)}
				{review.spoiler && <span className="text-muted">{t("reviews.spoilerTag")}</span>}
			</header>
			{shown ? (
				<Markdown nodes={review.body} />
			) : (
				<button type="button" className="underline" onClick={() => setShown(true)}>
					{t("reviews.reveal")}
				</button>
			)}
		</article>
	);
}

export default function Reviews({ itemId, reviews, canWrite, rated }: Props) {
	const t = useT();

	return (
		<div className="space-y-6">
			{canWrite &&
				(rated ? (
					<ReviewForm
						key={reviews.mine?.updatedAt ?? "new"}
						itemId={itemId}
						source={reviews.mine?.source}
						spoiler={reviews.mine?.spoiler}
					/>
				) : (
					<p className="text-muted">{t("reviews.rateFirst")}</p>
				))}
			{reviews.data.length === 0 ? (
				<p>{t("reviews.empty")}</p>
			) : (
				<ul className="divide-y divide-border">
					{reviews.data.map((r) => (
						<li key={r.id} className="space-y-2 pb-3">
							<ReviewCard review={r} />
							<ReactionBar
								target="review"
								id={r.id}
								likes={r.likes ?? 0}
								helpful={r.helpful}
								mine={r.myReaction}
								canReact={canWrite && r.source === undefined}
							/>
							<CommentThread
								target="review"
								id={r.id}
								comments={r.comments ?? []}
								canWrite={canWrite}
							/>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
