import type { ReactNode } from "react";
import { CopyButton } from "@/components/api/copy-button";

// Building blocks of the /docs pages. Technical content is English-only, so labels here are not translated.

export function DocTitle({ children, lead }: { children: ReactNode; lead?: ReactNode }) {
	return (
		<header className="mb-8 space-y-3">
			<h1 className="font-serif text-3xl font-bold sm:text-4xl">{children}</h1>
			{lead && <p className="max-w-prose text-lg text-muted">{lead}</p>}
		</header>
	);
}

export function H2({ id, children }: { id: string; children: ReactNode }) {
	return (
		<h2 id={id} className="mt-10 mb-3 scroll-mt-20 font-serif text-2xl font-bold">
			<a href={`#${id}`} className="hover:underline">
				{children}
			</a>
		</h2>
	);
}

export function P({ children }: { children: ReactNode }) {
	return <p className="my-3 max-w-prose leading-7">{children}</p>;
}

export function C({ children }: { children: ReactNode }) {
	return <code className="rounded bg-surface px-1.5 py-0.5 font-mono text-[0.9em]">{children}</code>;
}

export function Code({ children, title }: { children: string; title?: string }) {
	return (
		<figure className="my-4 overflow-hidden rounded-lg border border-border">
			<figcaption className="flex items-center justify-between gap-2 border-b border-border bg-surface px-3 py-1.5 text-xs text-muted">
				<span>{title ?? ""}</span>
				<CopyButton text={children} label="Copy" copiedLabel="Copied" />
			</figcaption>
			<pre className="overflow-x-auto p-4 font-mono text-sm leading-6">
				<code>{children}</code>
			</pre>
		</figure>
	);
}

export function Note({ children }: { children: ReactNode }) {
	return <aside className="my-4 rounded-md border border-border bg-surface p-4 text-sm leading-6">{children}</aside>;
}

export function Table({ head, rows }: { head: string[]; rows: ReactNode[][] }) {
	return (
		<div className="my-4 overflow-x-auto">
			<table className="w-full border-collapse text-left text-sm">
				<thead>
					<tr className="border-b border-border">
						{head.map((h) => (
							<th key={h} scope="col" className="py-2 pr-4 font-semibold">
								{h}
							</th>
						))}
					</tr>
				</thead>
				<tbody>
					{rows.map((row, i) => (
						// biome-ignore lint/suspicious/noArrayIndexKey: static rows, never reordered
						<tr key={i} className="border-b border-border align-top">
							{row.map((cell, j) => (
								// biome-ignore lint/suspicious/noArrayIndexKey: static cells, never reordered
								<td key={j} className="py-2 pr-4">
									{cell}
								</td>
							))}
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}
