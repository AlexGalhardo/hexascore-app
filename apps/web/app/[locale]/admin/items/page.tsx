import { Link } from "@/i18n/navigation";
import { findItems } from "@/lib/admin";
import { db } from "@/lib/db/client";
import { button, cell, head, input, requireAdmin } from "../guard";

type Props = { searchParams: Promise<{ q?: string }> };

export default async function AdminItemsPage({ searchParams }: Props) {
	await requireAdmin();
	const raw = (await searchParams).q;
	const q = typeof raw === "string" ? raw.slice(0, 200) : "";
	const results = await findItems(db, q);
	return (
		<section className="grid gap-4">
			<h1 className="font-serif text-2xl font-bold">Items</h1>
			<form className="flex flex-wrap gap-2">
				<input
					name="q"
					defaultValue={q}
					placeholder="id, slug or English title"
					aria-label="Search items"
					className={`${input} w-72`}
				/>
				<button type="submit" className={button}>
					Search
				</button>
			</form>
			{q ? (
				<div className="overflow-x-auto">
					<table className="w-full min-w-[36rem] text-sm">
						<thead>
							<tr>
								<th className={head}>Title</th>
								<th className={head}>Type</th>
								<th className={head}>Slug</th>
								<th className={`${head} text-right`}>Ratings</th>
								<th className={head}>Id</th>
							</tr>
						</thead>
						<tbody>
							{results.map((item) => (
								<tr key={item.id}>
									<td className={cell}>
										<Link href={`/admin/items/${item.id}`} className="underline">
											{item.title}
										</Link>
									</td>
									<td className={cell}>{item.type}</td>
									<td className={cell}>{item.slug}</td>
									<td className={`${cell} text-right font-mono tabular-nums`}>{item.ratingCount}</td>
									<td className={`${cell} font-mono text-xs`}>{item.id}</td>
								</tr>
							))}
						</tbody>
					</table>
					{results.length ? null : <p className="py-4 text-muted">No items found.</p>}
				</div>
			) : null}
		</section>
	);
}
