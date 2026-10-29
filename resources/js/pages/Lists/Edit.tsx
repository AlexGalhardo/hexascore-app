import { Link, router, useForm, usePage } from "@inertiajs/react";
import { type FormEvent, useState } from "react";
import Field from "../../components/Field";
import ItemLink, { type ItemRef } from "../../components/library/ItemLink";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Entry {
	position: number;
	note: string | null;
	item: ItemRef;
}

interface Props {
	list: {
		id: string;
		title: string;
		description: string | null;
		isPublic: boolean;
		hidden: boolean;
	};
	items: Entry[];
	version: number;
}

const button = "rounded-md border border-border px-3 py-1 text-sm disabled:opacity-50";

export default function Edit({ list, items, version }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const [order, setOrder] = useState(items);
	const details = useForm({
		title: list.title,
		description: list.description ?? "",
		isPublic: list.isPublic,
	});
	const base = `/lists/${list.id}`;
	const dirty = order.some((e, i) => e.item.id !== items[i]?.item.id);

	const move = (from: number, to: number) => {
		const next = [...order];
		const [moved] = next.splice(from, 1);
		next.splice(to, 0, moved);
		setOrder(next);
	};

	const saveDetails = (e: FormEvent) => {
		e.preventDefault();
		details.patch(base);
	};

	return (
		<>
			<Seo title={t("userLists.edited")} noindex />
			<Link href={localized("/lists", locale)} className="text-sm underline">
				{t("userLists.back")}
			</Link>
			<h1 className="mt-2 text-2xl font-semibold">{t("userLists.edited")}</h1>
			<form onSubmit={saveDetails} className="mt-4 max-w-md space-y-3">
				<Field
					label={t("userLists.titleField")}
					name="title"
					value={details.data.title}
					onChange={(e) => details.setData("title", e.target.value)}
					error={details.errors.title}
					required
					maxLength={80}
				/>
				<label className="block">
					<span className="font-medium">{t("userLists.description")}</span>
					<textarea
						value={details.data.description}
						onChange={(e) => details.setData("description", e.target.value)}
						maxLength={500}
						rows={3}
						className="mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700"
					/>
				</label>
				<label className="flex items-center gap-2 text-sm">
					<input
						type="checkbox"
						checked={details.data.isPublic}
						onChange={(e) => details.setData("isPublic", e.target.checked)}
					/>
					{t("userLists.public")}
				</label>
				<div className="flex gap-3">
					<button type="submit" disabled={details.processing} className={button}>
						{t("userLists.save")}
					</button>
					<button
						type="button"
						className={button}
						onClick={() => {
							if (window.confirm(t("userLists.confirmDelete"))) router.delete(base);
						}}
					>
						{t("userLists.delete")}
					</button>
				</div>
			</form>
			<h2 className="mt-8 text-lg font-semibold">{t("lists.addToList")}</h2>
			{order.length === 0 ? (
				<p className="mt-2">{t("userLists.empty")}</p>
			) : (
				<>
					<p className="mt-1 text-sm text-muted">{t("userLists.reorderHelp")}</p>
					<ol className="mt-2 divide-y divide-border">
						{order.map((e, i) => (
							<li key={e.item.id} className="flex flex-wrap items-center gap-2 py-2">
								<span className="w-8 text-sm text-muted">{i + 1}</span>
								<span className="min-w-0 flex-1">
									<ItemLink item={e.item} />
								</span>
								<button
									type="button"
									className={button}
									disabled={i === 0}
									aria-label={`${t("userLists.moveUp")}: ${e.item.title}`}
									onClick={() => move(i, i - 1)}
								>
									{t("userLists.moveUp")}
								</button>
								<button
									type="button"
									className={button}
									disabled={i === order.length - 1}
									aria-label={`${t("userLists.moveDown")}: ${e.item.title}`}
									onClick={() => move(i, i + 1)}
								>
									{t("userLists.moveDown")}
								</button>
								<button
									type="button"
									className={button}
									aria-label={`${t("userLists.remove")}: ${e.item.title}`}
									onClick={() => router.delete(`${base}/items/${e.item.id}`)}
								>
									{t("userLists.remove")}
								</button>
							</li>
						))}
					</ol>
					<button
						type="button"
						className={`${button} mt-3`}
						disabled={!dirty}
						onClick={() =>
							router.put(`${base}/order`, {
								version,
								itemIds: order.map((e) => e.item.id),
							})
						}
					>
						{t("userLists.saveOrder")}
					</button>
				</>
			)}
		</>
	);
}
