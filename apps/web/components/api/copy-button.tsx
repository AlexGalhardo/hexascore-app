"use client";

import { useState } from "react";

export function CopyButton({ text, label, copiedLabel }: { text: string; label: string; copiedLabel: string }) {
	const [copied, setCopied] = useState(false);

	async function copy(): Promise<void> {
		try {
			await navigator.clipboard.writeText(text);
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		} catch {
			// Clipboard can be blocked (permissions, insecure context); the text stays selectable on screen.
			setCopied(false);
		}
	}

	return (
		<button
			type="button"
			onClick={copy}
			className="min-h-9 rounded-md border border-border bg-bg px-3 text-xs hover:bg-surface"
		>
			<span aria-live="polite">{copied ? copiedLabel : label}</span>
		</button>
	);
}
