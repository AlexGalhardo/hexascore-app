import type { ReactNode } from "react";

/** One node of a review parsed on the server (SDD 14 §12.10): text only ever reaches the DOM as React text. */
export interface MdNode {
	t: "p" | "br" | "strong" | "em" | "code" | "a" | "ul" | "ol" | "li" | "blockquote" | "text";
	text?: string;
	href?: string;
	c?: MdNode[];
}

const safeHref = (href = ""): boolean => /^https?:\/\//i.test(href);

function node(n: MdNode, key: number): ReactNode {
	const kids = n.c?.map(node);
	switch (n.t) {
		case "text":
			return n.text;
		case "br":
			return <br key={key} />;
		case "code":
			return (
				<code key={key} className="rounded-xs bg-surface px-1">
					{n.text}
				</code>
			);
		case "a":
			// The server already allows only http(s); checked again so a bad payload never becomes a live link.
			return safeHref(n.href) ? (
				<a key={key} href={n.href} rel="nofollow ugc noopener" className="underline">
					{kids}
				</a>
			) : (
				<span key={key}>{kids}</span>
			);
		case "ul":
			return (
				<ul key={key} className="list-disc pl-6">
					{kids}
				</ul>
			);
		case "ol":
			return (
				<ol key={key} className="list-decimal pl-6">
					{kids}
				</ol>
			);
		case "blockquote":
			return (
				<blockquote key={key} className="border-l-2 border-border pl-3 text-muted">
					{kids}
				</blockquote>
			);
		case "p":
		case "strong":
		case "em":
		case "li": {
			const Tag = n.t;
			return <Tag key={key}>{kids}</Tag>;
		}
		default:
			return null;
	}
}

export default function Markdown({ nodes }: { nodes: MdNode[] }) {
	return <div className="space-y-2">{nodes.map(node)}</div>;
}
