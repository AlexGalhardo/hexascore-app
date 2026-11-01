import { router } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import Btn from "../ui/Btn";
import ReactionBar, { type ReactionKind } from "./ReactionBar";

const MAX = 2000;

/** One comment of a thread (SDD 14 §12.11): plain text, replies only under a top-level comment. */
export interface CommentEntry {
	id: string;
	author: { nickname: string; initial: string };
	body: string;
	likes: number;
	helpful: number;
	myReaction?: ReactionKind;
	mine?: boolean;
	createdAt: string;
	replies?: CommentEntry[];
}

interface Props {
	target: "review" | "list";
	id: string;
	comments: CommentEntry[];
	/** A signed-in reader may comment, reply and react. */
	canWrite: boolean;
}

function CommentForm({
	url,
	parentId = "",
	label,
	onDone,
}: {
	url: string;
	parentId?: string;
	label: string;
	onDone?: () => void;
}) {
	const t = useT();
	const [body, setBody] = useState("");

	const send = (e: FormEvent) => {
		e.preventDefault();
		router.post(
			url,
			{ body, parentId },
			{
				preserveScroll: true,
				onSuccess: () => {
					setBody("");
					onDone?.();
				},
			},
		);
	};

	return (
		<form onSubmit={send} className="flex flex-wrap items-end gap-2">
			<label className="min-w-0 flex-1">
				<span className="sr-only">{label}</span>
				<textarea
					aria-label={label}
					value={body}
					maxLength={MAX}
					rows={2}
					required
					onChange={(e) => setBody(e.target.value)}
					className="block w-full rounded-xs border border-border bg-bg p-2"
				/>
			</label>
			<Btn type="submit" disabled={body.trim() === ""}>
				{t("comments.send")}
			</Btn>
		</form>
	);
}

/** A comment thread under a review or a list, with reactions on each comment. */
export default function CommentThread({ target, id, comments, canWrite }: Props) {
	const t = useT();
	const l = useLocalized();
	const [replyTo, setReplyTo] = useState("");
	const url = l(`/comments/${target}/${id}`);

	const item = (c: CommentEntry, top: boolean) => (
		<li key={c.id} className="space-y-1 py-2">
			<p className="text-sm font-semibold">{c.author.nickname}</p>
			<p className="whitespace-pre-line">{c.body}</p>
			<div className="flex flex-wrap items-center gap-3 text-sm">
				<ReactionBar
					target="comment"
					id={c.id}
					likes={c.likes}
					helpful={c.helpful}
					mine={c.myReaction}
					canReact={canWrite && !c.mine}
				/>
				{top && canWrite && (
					<button
						type="button"
						className="min-h-11 underline"
						onClick={() => setReplyTo(c.id)}
					>
						{t("comments.reply")}
					</button>
				)}
				{c.mine && (
					<button
						type="button"
						className="min-h-11 underline"
						onClick={() =>
							router.delete(l(`/comments/${c.id}`), { preserveScroll: true })
						}
					>
						{t("comments.delete")}
					</button>
				)}
			</div>
			{top && replyTo === c.id && (
				<CommentForm
					url={url}
					parentId={c.id}
					label={t("comments.replyLabel")}
					onDone={() => setReplyTo("")}
				/>
			)}
			{c.replies && c.replies.length > 0 && (
				<ul className="ml-6 border-l border-border pl-4">
					{c.replies.map((r) => item(r, false))}
				</ul>
			)}
		</li>
	);

	return (
		<section aria-label={t("comments.title")} className="space-y-2">
			{comments.length > 0 && <ul>{comments.map((c) => item(c, true))}</ul>}
			{canWrite && <CommentForm url={url} label={t("comments.label")} />}
		</section>
	);
}
