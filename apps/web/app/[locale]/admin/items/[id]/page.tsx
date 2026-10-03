import { notFound } from "next/navigation";
import { deleteAffiliateLink, mergeItems, saveAffiliateLink } from "@/actions/admin";
import { ActionForm } from "@/components/admin/action-form";
import { CategoryLabel } from "@/components/catalog/category-label";
import { CoverThumb } from "@/components/catalog/cover-thumb";
import { Link } from "@/i18n/navigation";
import { AFFILIATE_STORES, getAdminItem } from "@/lib/admin";
import { db } from "@/lib/db/client";
import { itemHref } from "@/lib/sections";
import { button, cell, head, input, requireAdmin } from "../../guard";

type Props = { params: Promise<{ id: string }> };

export default async function AdminItemPage({ params }: Props) {
	await requireAdmin();
	const item = await getAdminItem(db, (await params).id);
	if (!item) notFound();
	return (
		<section className="grid gap-6">
			<header className="flex items-center gap-3">
				<CoverThumb src={item.coverUrl} thumb={item.coverThumbUrl} size="md" />
				<div className="grid min-w-0 gap-1">
					<h1 className="font-serif text-2xl font-bold">{item.title}</h1>
					<p className="text-sm text-muted">
						<CategoryLabel type={item.type}>{item.type}</CategoryLabel> · {item.slug} · {item.ratingCount}{" "}
						ratings · <code>{item.id}</code> ·{" "}
						<Link href={itemHref(item.type, item.slug)} className="underline">
							public page
						</Link>
					</p>
				</div>
			</header>

			<section className="grid gap-2">
				<h2 className="font-serif text-lg font-bold">Affiliate links</h2>
				<p className="text-xs text-muted">https only; the host must belong to the store (validated on save).</p>
				<div className="overflow-x-auto">
					<table className="w-full min-w-[36rem] text-sm">
						<thead>
							<tr>
								<th className={head}>Store</th>
								<th className={head}>URL</th>
								<th className={head}>Active</th>
								<th className={head} />
							</tr>
						</thead>
						<tbody>
							{item.affiliates.map((a) => (
								<tr key={a.store}>
									<td className={cell}>{a.store}</td>
									<td className={`${cell} max-w-sm break-all`}>{a.url}</td>
									<td className={cell}>{a.active ? "yes" : "no"}</td>
									<td className={cell}>
										<ActionForm action={deleteAffiliateLink} confirm="Delete this link?">
											<input type="hidden" name="itemId" value={item.id} />
											<input type="hidden" name="store" value={a.store} />
											<button type="submit" className={button}>
												Delete
											</button>
										</ActionForm>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
				<ActionForm action={saveAffiliateLink}>
					<input type="hidden" name="itemId" value={item.id} />
					<select name="store" aria-label="Store" className={input}>
						{AFFILIATE_STORES.map((s) => (
							<option key={s} value={s}>
								{s}
							</option>
						))}
					</select>
					<input
						name="url"
						type="url"
						required
						placeholder="https://…"
						aria-label="URL"
						className={`${input} w-80`}
					/>
					<label className="flex items-center gap-1 text-sm">
						<input type="checkbox" name="active" defaultChecked /> active
					</label>
					<button type="submit" className={button}>
						Save link
					</button>
				</ActionForm>
			</section>

			<section className="grid gap-2">
				<h2 className="font-serif text-lg font-bold">Merge duplicate</h2>
				<p className="text-xs text-muted">
					Moves ratings, statuses, list entries, affiliate links, external scores and source records from the
					source into this item, then deletes the source. When a user has rated both, this item&apos;s rating
					is kept. Same type only.
				</p>
				<ActionForm action={mergeItems} confirm="Merge and delete the source item? This cannot be undone.">
					<input type="hidden" name="targetId" value={item.id} />
					<input
						name="sourceId"
						required
						placeholder="source item id (deleted)"
						aria-label="Source item id"
						className={`${input} w-80 font-mono`}
					/>
					<button type="submit" className={`${button} text-danger`}>
						Merge into this item
					</button>
				</ActionForm>
			</section>
		</section>
	);
}
