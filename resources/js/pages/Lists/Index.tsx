import { Link, useForm, usePage } from "@inertiajs/react";
import type { FormEvent } from "react";

import Field, { fieldClass } from "../../components/Field";
import Seo from "../../components/Seo";
import SubmitButton from "../../components/SubmitButton";
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
		form.post(localized("/lists", locale), { onSuccess: () => form.reset("title") });
	};

	return (
		<>
			<Seo title={t("userLists.title")} noindex />
			<h1 className="mt-6 font-display text-5xl leading-none font-bold tracking-[-0.02em] md:mt-10 md:text-[64px]">
				{t("userLists.title")}
			</h1>
			<p className="mt-3 text-sm text-muted">
				{t("userLists.limit", { limit: limits.perType })}
			</p>
			{lists.length === 0 ? (
				<p className="mt-8 border-t-2 border-fg pt-6 font-display text-2xl italic">
					{t("lists.noLists")}
				</p>
			) : (
				<ul className="mt-8 divide-y divide-border border-t-2 border-fg">
					{lists.map((l) => (
						<li key={l.id} className="flex flex-wrap items-center gap-x-4 gap-y-1 py-3">
							<Link
								href={localized(`/lists/${l.id}/edit`, locale)}
								className="font-display text-xl font-bold hover:text-primary"
							>
								{l.title}
							</Link>
							<span className="text-sm text-muted">
								{t(`itemTypes.${l.type}`)} ·{" "}
								{t("userLists.itemCount", { count: l.itemCount })} ·{" "}
								{l.isPublic ? t("userLists.public") : t("userLists.private")}
							</span>
							{l.hidden && (
								<span className="text-sm text-danger">
									{t("userLists.hiddenByMod")}
								</span>
							)}
						</li>
					))}
				</ul>
			)}
			<form onSubmit={submit} className="mt-10 max-w-md space-y-5 border-t-2 border-fg pt-4">
				<h2 className="font-display text-[28px] leading-tight font-bold">
					{t("userLists.new")}
				</h2>
				<label className="block">
					<span className="text-sm font-semibold">{t("lists.fieldType")}</span>
					<select
						value={form.data.type}
						onChange={(e) => form.setData("type", e.target.value)}
						className={fieldClass}
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
				<label className="flex min-h-11 items-center gap-2.5">
					<input
						type="checkbox"
						checked={form.data.isPublic}
						onChange={(e) => form.setData("isPublic", e.target.checked)}
						className="size-5 accent-fg"
					/>
					{t("userLists.public")}
				</label>
				<SubmitButton pending={form.processing}>{t("userLists.create")}</SubmitButton>
			</form>
		</>
	);
}
