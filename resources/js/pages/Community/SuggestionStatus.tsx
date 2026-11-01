import { Link, router } from "@inertiajs/react";
import Seo from "../../components/Seo";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

interface Change {
	before?: unknown;
	after?: unknown;
	add?: { id: string; role?: string }[];
	remove?: { id: string; role?: string }[];
}

interface Suggestion {
	id: string;
	kind: string;
	targetType: string;
	targetId: string | null;
	domain: string;
	status: string;
	version: number;
	note: string;
	sourceUrls: string[];
	changes: Record<string, Change>;
	names?: Record<string, string>;
	rejectReason: string | null;
	decisionNote: string | null;
	events: {
		from: string | null;
		to: string;
		actorRole: string;
		reason: string | null;
		at: string;
	}[];
}

interface Props {
	suggestion: Suggestion;
	canAct: boolean;
	quota: { limit: number; used: number; left: number; resetAt: string };
	open: boolean;
}

const show = (v: unknown): string => {
	if (v === null || v === undefined || v === "") return "∅";
	return typeof v === "string" ? v : JSON.stringify(v);
};

// Text markers ("-" removed, "+" added) so the diff never relies on colour alone (NFR-04-6).
function diffLines(c: Change, names: Record<string, string>): { sign: "-" | "+"; text: string }[] {
	const out: { sign: "-" | "+"; text: string }[] = [];
	const named = (v: unknown) =>
		typeof v === "string" && names[v] ? `${names[v]} (${v})` : show(v);
	if ("after" in c || "before" in c) {
		out.push({ sign: "-", text: named(c.before) }, { sign: "+", text: named(c.after) });
	}
	const label = (o: { id: string; role?: string }) =>
		[names[o.id] ? `${names[o.id]} (${o.id})` : o.id, o.role].filter(Boolean).join(" ");
	for (const r of c.remove ?? []) out.push({ sign: "-", text: label(r) });
	for (const a of c.add ?? []) out.push({ sign: "+", text: label(a) });
	return out;
}

const post = (url: string, version: number) =>
	router.post(url, { version }, { preserveScroll: true });

export default function SuggestionStatus({ suggestion: s, canAct, quota, open }: Props) {
	const t = useT();
	const l = useLocalized();
	const fmt = (iso: string) => new Date(iso).toLocaleString();
	const canSubmit = canAct && open && s.status === "draft" && quota.left > 0;
	const canWithdraw = canAct && ["draft", "submitted", "in_review"].includes(s.status);

	return (
		<>
			<Seo title={t("community.status.title")} noindex />
			<h1 className="text-2xl font-semibold">{t("community.status.title")}</h1>
			<p className="mt-2" data-testid="status">
				{t("community.status.current")}:{" "}
				<strong>{t(`community.status.states.${s.status}`)}</strong>
			</p>
			{s.status === "rejected" && s.rejectReason && (
				<p className="mt-1">
					{t("community.status.reason")}: {s.rejectReason}
					{s.decisionNote ? ` (${s.decisionNote})` : ""}
				</p>
			)}
			{canAct && s.status === "draft" && quota.left === 0 && (
				<p role="status" className="mt-1">
					{t("community.status.quotaReached", { time: fmt(quota.resetAt) })}
				</p>
			)}

			<h2 className="mt-6 text-lg font-semibold">{t("community.status.changes")}</h2>
			<dl className="mt-2 space-y-3">
				{Object.entries(s.changes).map(([field, c]) => (
					<div key={field}>
						<dt className="font-medium">{field}</dt>
						<dd>
							<ul className="font-mono text-sm">
								{diffLines(c, s.names ?? {}).map((l, i) => (
									// biome-ignore lint/suspicious/noArrayIndexKey: static, ordered diff lines
									<li key={i}>
										<span aria-hidden="true">{l.sign} </span>
										<span className="sr-only">
											{l.sign === "-"
												? t("community.status.removed")
												: t("community.status.added")}
											:
										</span>
										{l.text}
									</li>
								))}
							</ul>
						</dd>
					</div>
				))}
			</dl>
			{s.note && <p className="mt-3 text-sm">{s.note}</p>}

			<h2 className="mt-6 text-lg font-semibold">{t("community.status.timeline")}</h2>
			<ol className="mt-2 space-y-1 text-sm">
				{s.events.map((e) => (
					<li key={`${e.at}-${e.to}`}>
						{fmt(e.at)}: {e.from ? `${t(`community.status.states.${e.from}`)} → ` : ""}
						{t(`community.status.states.${e.to}`)} ({e.actorRole})
						{e.reason ? `: ${e.reason}` : ""}
					</li>
				))}
			</ol>

			<div className="mt-6 flex gap-3">
				{canSubmit && (
					<button
						type="button"
						onClick={() => post(l(`/suggestions/${s.id}/submit`), s.version)}
						className="rounded-md border border-border px-3 py-1 text-sm"
					>
						{t("community.status.submit")}
					</button>
				)}
				{canWithdraw && (
					<button
						type="button"
						onClick={() => post(l(`/suggestions/${s.id}/withdraw`), s.version)}
						className="rounded-md border border-border px-3 py-1 text-sm"
					>
						{t("community.status.withdraw")}
					</button>
				)}
				<Link href={l("/me/suggestions")} className="px-3 py-1 text-sm underline">
					{t("community.mine.title")}
				</Link>
			</div>
		</>
	);
}
