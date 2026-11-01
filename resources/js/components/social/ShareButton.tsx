import { useState } from "react";

import { useT } from "../../lib/useT";
import Btn from "../ui/Btn";

/** Shares the current page: the Web Share API when the browser has it, else the link goes to the clipboard. */
export default function ShareButton({ title }: { title: string }): React.JSX.Element {
	const t = useT();
	const [copied, setCopied] = useState(false);

	async function share(): Promise<void> {
		const url = window.location.href;
		if (navigator.share) {
			// A dismissed share sheet rejects with AbortError; nothing to report.
			await navigator.share({ title, url }).catch(() => undefined);
			return;
		}
		await navigator.clipboard.writeText(url);
		setCopied(true);
	}

	return (
		<span className="inline-flex items-center gap-2 text-sm">
			<Btn kind="secondary" onClick={() => void share()}>
				{t("social.share.action")}
			</Btn>
			<span role="status" className="text-muted">
				{copied ? t("social.share.copied") : ""}
			</span>
		</span>
	);
}
