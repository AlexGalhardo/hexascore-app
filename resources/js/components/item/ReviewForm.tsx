import { router } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import Btn from "../ui/Btn";

const MAX = 5000;

interface Props {
	itemId: string;
	source?: string;
	spoiler?: boolean;
}

/** Write or edit the viewer's review of a rated item (SDD 14 §12.10). */
export default function ReviewForm({ itemId, source = "", spoiler = false }: Props) {
	const t = useT();
	const l = useLocalized();
	const [body, setBody] = useState(source);
	const [isSpoiler, setSpoiler] = useState(spoiler);
	const url = l(`/items/${itemId}/review`);
	const options = { preserveScroll: true };

	const save = (e: FormEvent) => {
		e.preventDefault();
		router.put(url, { body, spoiler: isSpoiler }, options);
	};

	return (
		<form onSubmit={save} className="space-y-3">
			<label className="block">
				<span className="font-medium">{t("reviews.body")}</span>
				<textarea
					value={body}
					maxLength={MAX}
					rows={5}
					required
					onChange={(e) => setBody(e.target.value)}
					className="mt-1 block w-full rounded-xs border border-border bg-bg p-2"
					aria-describedby="review-help"
				/>
			</label>
			<p id="review-help" className="text-sm text-muted">
				{t("reviews.help")}
			</p>
			<label className="flex items-center gap-2">
				<input
					type="checkbox"
					checked={isSpoiler}
					onChange={(e) => setSpoiler(e.target.checked)}
				/>
				<span>{t("reviews.spoiler")}</span>
			</label>
			<div className="flex flex-wrap gap-3">
				<Btn type="submit" disabled={body.trim() === ""}>
					{t("reviews.save")}
				</Btn>
				{source && (
					<Btn kind="secondary" onClick={() => router.delete(url, options)}>
						{t("reviews.delete")}
					</Btn>
				)}
			</div>
		</form>
	);
}
