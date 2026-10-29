import { Link, useForm, usePage } from "@inertiajs/react";
import type { FormEvent } from "react";

import Field from "../../components/Field";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const TYPES = ["game", "movie", "series", "book"] as const;

interface ListRow {
	id: string;
	type: (typeof TYPES)[number];
	title: string;
	isPublic: boolean;
	hidden: boolean;
	itemCount: number;
}

interface Props {
	lists: ListRow[];
	limits: { perType: number; items: number };
}

export default function Index({ lists, limits }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const form = useForm({ type: "game", title: "", isPublic: true });

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post("/lists", { onSuccess: () => form.reset("title") });
	};

	return (
		<>
			<Seo title={t("userLists.title")} noindex />
			<h1 className="text-2xl font-semibold">{t("userLists.title")}</h1>
			<p className="mt-1 text-sm text-muted">
				{t("userLists.limit", { limit: limits.perType })}
			</p>
			{lists.length === 0 ? (
				<p className="mt-4">{t("lists.noLists")}</p>
			) : (
				<ul className="mt-4 divide-y divide-border">
					{lists.map((l) => (
						<li key={l.id} className="flex flex-wrap items-center gap-x-4 gap-y-1 py-3">
							<Link
								href={localized(`/lists/${l.id}/edit`, locale)}
								className="font-medium underline"
							>
								{l.title}
							</Link>
							<span className="text-sm text-muted">
								{t(`itemTypes.${l.type}`)} ·{" "}
								{t("userLists.itemCount", { count: l.itemCount })} ·{" "}
								{l.isPublic ? t("userLists.public") : t("userLists.private")}
							</span>
							{l.hidden && (
								<span className="text-sm text-red-600">
									{t("userLists.hiddenByMod")}
								</span>
							)}
						</li>
					))}
				</ul>
			)}
			<form onSubmit={submit} className="mt-8 max-w-md space-y-3">
				<h2 className="text-lg font-semibold">{t("userLists.new")}</h2>
				<label className="flex flex-col text-sm">
					{t("lists.fieldType")}
					<select
						value={form.data.type}
						onChange={(e) => form.setData("type", e.target.value)}
						className="rounded-md border border-border bg-surface px-2 py-1"
					>
						{TYPES.map((x) => (
							<option key={x} value={x}>
								{t(`itemTypes.${x}`)}
							</option>
						))}
					</select>
				</label>
				<Field
					label={t("userLists.titleField")}
					name="title"
					value={form.data.title}
					onChange={(e) => form.setData("title", e.target.value)}
					error={form.errors.title}
					required
					maxLength={80}
				/>
				<label className="flex items-center gap-2 text-sm">
					<input
						type="checkbox"
						checked={form.data.isPublic}
						onChange={(e) => form.setData("isPublic", e.target.checked)}
					/>
					{t("userLists.public")}
				</label>
				<button
					type="submit"
					disabled={form.processing}
					className="rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
				>
					{t("userLists.create")}
				</button>
			</form>
		</>
	);
}
