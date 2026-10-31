import { Link, router, usePage } from "@inertiajs/react";
import { type FormEvent, useState } from "react";
import FieldEditor from "../../components/community/FieldEditor";
import { asText, editable, fieldLabel } from "../../components/community/fields";
import Seo from "../../components/Seo";
import { useT } from "../../lib/useT";

interface Draft {
	id: string;
	version: number;
	locale: string;
	note: string;
	sourceUrls: string[];
	changes: Record<string, { before?: unknown; after?: unknown }>;
}

interface Props {
	open: boolean;
	domain?: string;
	kind?: "create" | "merge";
	fields?: string[];
	quota?: { limit: number; used: number; left: number };
	draft?: Draft | null;
}

const inputClass =
	"mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700";

export default function SuggestCreate({
	open,
	domain = "",
	kind = "create",
	fields = [],
	quota,
	draft,
}: Props) {
	const t = useT();
	const { errors } = usePage<{ errors?: Record<string, string> }>().props;
	const [edits, setEdits] = useState<Record<string, string>>(() =>
		Object.fromEntries(
			Object.entries(draft?.changes ?? {}).map(([k, c]) => [k, asText(c.after)]),
		),
	);
	const [sourceId, setSourceId] = useState(asText(draft?.changes["*merge*"]?.before));
	const [destId, setDestId] = useState(asText(draft?.changes["*merge*"]?.after));
	const [locale, setLocale] = useState(draft?.locale ?? "en");
	const [note, setNote] = useState(draft?.note ?? "");
	const [urls, setUrls] = useState((draft?.sourceUrls ?? []).join("\n"));
	const [website, setWebsite] = useState("");

	if (!open) {
		return (
			<>
				<Seo title={t("community.create.title")} noindex />
				<h1 className="text-2xl font-semibold">{t("community.closed.title")}</h1>
				<p className="mt-2">{t("community.closed.body")}</p>
			</>
		);
	}

	const submit = (e: FormEvent) => {
		e.preventDefault();
		const changes: Record<string, { before: string | null; after: string }> = {};
		if (kind === "merge") {
			changes["*merge*"] = { before: sourceId.trim(), after: destId.trim() };
		} else {
			for (const [f, after] of Object.entries(edits)) {
				if (after.trim() !== "") changes[f] = { before: null, after };
			}
		}
		const body = {
			kind,
			domain,
			targetKind: "item",
			targetId: kind === "merge" ? sourceId.trim() : "",
			version: draft?.version ?? 0,
			locale,
			note,
			sourceUrls: urls
				.split("\n")
				.map((u) => u.trim())
				.filter(Boolean),
			changes,
			website,
		};
		if (draft) router.patch(`/suggestions/${draft.id}`, body, { preserveScroll: true });
		else router.post("/suggestions", body, { preserveScroll: true });
	};

	const errorList = Object.entries(errors ?? {});
	const tab = (k: "create" | "merge", text: string) => (
		<Link
			href={`/suggest/new/${domain}?kind=${k}`}
			aria-current={kind === k ? "page" : undefined}
			className={`rounded-md border px-3 py-1 text-sm ${kind === k ? "border-text font-medium" : "border-border"}`}
		>
			{text}
		</Link>
	);

	return (
		<>
			<Seo title={t("community.create.title")} noindex />
			<h1 className="text-2xl font-semibold">{t("community.create.title")}</h1>
			<nav aria-label={t("community.create.title")} className="mt-3 flex gap-2">
				{tab("create", t("community.create.newRecord"))}
				{tab("merge", t("community.create.reportDuplicate"))}
			</nav>
			{quota && (
				<p data-testid="quota-left" className="mt-2 text-sm text-muted">
					{t("community.quota", { left: quota.left, limit: quota.limit })}
				</p>
			)}
			{draft && <p className="mt-1 text-sm">{t("community.draftRestored")}</p>}
			{errorList.length > 0 && (
				<ul role="alert" className="mt-3 text-sm text-red-700 dark:text-red-400">
					{errorList.map(([f, m]) => (
						<li key={f}>
							{f}: {m}
						</li>
					))}
				</ul>
			)}
			<form onSubmit={submit} className="mt-4 max-w-xl space-y-4">
				{kind === "merge" ? (
					<>
						<p className="text-sm">{t("community.create.mergeHelp", { domain })}</p>
						<label className="block">
							<span className="font-medium">{t("community.create.mergeSource")}</span>
							<input
								value={sourceId}
								onChange={(e) => setSourceId(e.target.value)}
								required
								className={inputClass}
							/>
						</label>
						<label className="block">
							<span className="font-medium">{t("community.create.mergeDest")}</span>
							<input
								value={destId}
								onChange={(e) => setDestId(e.target.value)}
								required
								className={inputClass}
							/>
						</label>
					</>
				) : (
					fields
						.filter(editable)
						.map((f) => (
							<FieldEditor
								key={f}
								label={f === "title" ? `${fieldLabel(t, f)} *` : fieldLabel(t, f)}
								current=""
								value={edits[f] ?? ""}
								multiline={f.startsWith("synopsis.")}
								lockedLabel={t("community.locked")}
								onChange={(v) => setEdits((prev) => ({ ...prev, [f]: v }))}
							/>
						))
				)}
				<label className="block">
					<span className="font-medium">{t("community.locale")}</span>
					<select
						value={locale}
						onChange={(e) => setLocale(e.target.value)}
						className="mt-1 block rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700"
					>
						{["en", "pt-BR", "es"].map((l) => (
							<option key={l} value={l}>
								{l}
							</option>
						))}
					</select>
				</label>
				<label className="block">
					<span className="font-medium">{t("community.sources")}</span>
					<textarea
						value={urls}
						onChange={(e) => setUrls(e.target.value)}
						rows={3}
						className={inputClass}
					/>
				</label>
				<label className="block">
					<span className="font-medium">{t("community.note")}</span>
					<textarea
						value={note}
						onChange={(e) => setNote(e.target.value)}
						rows={3}
						maxLength={1000}
						className={inputClass}
					/>
				</label>
				<div aria-hidden="true" className="absolute -left-[9999px] h-0 w-0 overflow-hidden">
					<label>
						website
						<input
							type="text"
							name="website"
							tabIndex={-1}
							autoComplete="off"
							value={website}
							onChange={(e) => setWebsite(e.target.value)}
						/>
					</label>
				</div>
				<button type="submit" className="rounded-md border border-border px-3 py-1 text-sm">
					{t("community.saveDraft")}
				</button>
			</form>
		</>
	);
}
