import { getIngestionStatus } from "@/lib/admin";
import { db } from "@/lib/db/client";
import { cell, head, requireAdmin } from "../guard";

const fmt = (iso: string | null) => (iso ? iso.slice(0, 19).replace("T", " ") : "—");

export default async function AdminIngestionPage() {
	await requireAdmin();
	const sources = await getIngestionStatus(db);
	return (
		<section className="grid gap-4">
			<h1 className="font-serif text-2xl font-bold">Ingestion status</h1>
			<p className="text-sm text-muted">
				Per <code>source_records.source</code>; pending = never normalized or fetched again since (UTC).
			</p>
			<div className="overflow-x-auto">
				<table className="w-full min-w-[36rem] text-sm">
					<thead>
						<tr>
							<th className={head}>Source</th>
							<th className={`${head} text-right`}>Records</th>
							<th className={`${head} text-right`}>Pending</th>
							<th className={head}>Last fetched</th>
							<th className={head}>Last normalized</th>
						</tr>
					</thead>
					<tbody>
						{sources.map((s) => (
							<tr key={s.source}>
								<td className={cell}>{s.source}</td>
								<td className={`${cell} text-right font-mono tabular-nums`}>{s.records}</td>
								<td className={`${cell} text-right font-mono tabular-nums`}>{s.pending}</td>
								<td className={`${cell} font-mono text-xs`}>{fmt(s.lastFetchedAt)}</td>
								<td className={`${cell} font-mono text-xs`}>{fmt(s.lastNormalizedAt)}</td>
							</tr>
						))}
					</tbody>
				</table>
				{sources.length ? null : <p className="py-4 text-muted">No source records yet.</p>}
			</div>
		</section>
	);
}
