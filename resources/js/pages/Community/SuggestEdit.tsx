import { router, usePage } from "@inertiajs/react";
import { type FormEvent, useState } from "react";
import DateEditor from "../../components/community/DateEditor";
import FieldEditor from "../../components/community/FieldEditor";
import {
	asDate,
	asOps,
	asText,
	type Change,
	type DateValue,
	editable,
	idsText,
	isDate,
	isRelation,
	fieldLabel as label,
	parseIds,
	relationRoles,
} from "../../components/community/fields";
import RelationEditor, { type RelationDraft } from "../../components/community/RelationEditor";
import Seo from "../../components/Seo";
import { useT } from "../../lib/useT";

interface Lock {
	field: string;
	reason: string;
}

interface Draft {
	id: string;
	version: number;
	locale: string;
	note: string;
	sourceUrls: string[];
	changes: Record<string, { after?: unknown; add?: unknown; remove?: unknown }>;
}

interface Props {
	open: boolean;
	target?: { kind: string; id: string; domain: string; values: Record<string, unknown> };
	fields?: string[];
	locks?: Lock[];
	quota?: { limit: number; used: number; left: number };
	draft?: Draft | null;
}

export default function SuggestEdit({
	open,
	target,
	fields = [],
	locks = [],
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
	const [dates, setDates] = useState<Record<string, DateValue>>(() => {
		const after = asDate(draft?.changes.release_date?.after);
		const init: Record<string, DateValue> = {};
		if (after) init.release_date = after;
		return init;
	});
	const [rels, setRels] = useState<Record<string, RelationDraft>>(() =>
		Object.fromEntries(
			Object.entries(draft?.changes ?? {})
				.filter(([f]) => isRelation(f))
				.map(([f, c]) => {
					const add = asOps(c.add);
					return [
						f,
						{
							add: idsText(add),
							remove: idsText(asOps(c.remove)),
							role: add[0]?.role ?? "",
						},
					];
				}),
		),
	);
	const [locale, setLocale] = useState(draft?.locale ?? "en");
	const [note, setNote] = useState(draft?.note ?? "");
	const [urls, setUrls] = useState((draft?.sourceUrls ?? []).join("\n"));
	const [website, setWebsite] = useState("");

	if (!open || !target) {
		return (
			<>
				<Seo title={t("community.suggest.title")} noindex />
				<h1 className="text-2xl font-semibold">{t("community.closed.title")}</h1>
				<p className="mt-2">{t("community.closed.body")}</p>
			</>
		);
	}

	const lockFor = (f: string) => locks.find((l) => l.field === f || l.field === "*")?.reason;
	const submit = (e: FormEvent) => {
		e.preventDefault();
		const changes: Record<string, Change> = {};
		for (const [f, after] of Object.entries(edits)) {
			if (after !== asText(target.values[f]) && !lockFor(f)) {
				changes[f] = { before: asText(target.values[f]) || null, after };
			}
		}
		for (const [f, after] of Object.entries(dates)) {
			const before = asDate(target.values[f]);
			if (
				after.value !== "" &&
				(after.value !== before?.value || after.precision !== before?.precision) &&
				!lockFor(f)
			) {
				changes[f] = { before, after };
			}
		}
		for (const [f, r] of Object.entries(rels)) {
			const role = relationRoles[f]?.length ? r.role || relationRoles[f][0] : undefined;
			const add = parseIds(r.add, role);
			const remove = parseIds(r.remove, role);
			if (add.length + remove.length > 0 && !lockFor(f)) changes[f] = { add, remove };
		}
		const body = {
			targetKind: target.kind,
			targetId: target.id,
			domain: target.domain,
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

	return (
		<>
			<Seo title={t("community.suggest.title")} noindex />
			<h1 className="text-2xl font-semibold">{t("community.suggest.title")}</h1>
			{quota && (
				<p data-testid="quota-left" className="mt-1 text-sm text-muted">
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
				{fields.filter(editable).map((f) => (
					<FieldEditor
						key={f}
						label={label(t, f)}
						current={asText(target.values[f])}
						value={edits[f] ?? asText(target.values[f])}
						multiline={f.startsWith("synopsis.")}
						lockReason={lockFor(f)}
						lockedLabel={t("community.locked")}
						onChange={(v) => setEdits((prev) => ({ ...prev, [f]: v }))}
					/>
				))}
				{fields.filter(isDate).map((f) => (
					<DateEditor
						key={f}
						label={label(t, f)}
						current={asDate(target.values[f])}
						value={
							dates[f] ?? asDate(target.values[f]) ?? { value: "", precision: "day" }
						}
						lockReason={lockFor(f)}
						lockedLabel={t("community.locked")}
						onChange={(v) => setDates((prev) => ({ ...prev, [f]: v }))}
					/>
				))}
				{fields.filter(isRelation).map((f) => (
					<RelationEditor
						key={f}
						label={label(t, f)}
						roles={relationRoles[f]}
						value={rels[f] ?? { add: "", remove: "", role: "" }}
						lockReason={lockFor(f)}
						lockedLabel={t("community.locked")}
						addLabel={`${t("community.relation.add")} ${label(t, f)}`}
						removeLabel={`${t("community.relation.remove")} ${label(t, f)}`}
						roleLabel={`${t("community.relation.role")} ${label(t, f)}`}
						onChange={(v) => setRels((prev) => ({ ...prev, [f]: v }))}
					/>
				))}
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
						className="mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700"
					/>
				</label>
				<label className="block">
					<span className="font-medium">{t("community.note")}</span>
					<textarea
						value={note}
						onChange={(e) => setNote(e.target.value)}
						rows={3}
						maxLength={1000}
						className="mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700"
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
