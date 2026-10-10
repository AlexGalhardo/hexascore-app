import { getAdminOverview } from "@/lib/admin";
import { db } from "@/lib/db/client";
import { seedDataFlag } from "@/lib/flags";
import { requireAdmin } from "./guard";

export default async function AdminPage() {
	await requireAdmin();
	const [o, seedData] = await Promise.all([getAdminOverview(db), seedDataFlag()]);
	const rows: [string, number][] = [
		["Open reports", o.openReports],
		["Unhandled contact messages", o.unhandledMessages],
		["Hidden lists", o.hiddenLists],
		["Banned users", o.bannedUsers],
	];
	return (
		<section className="grid gap-4">
			<h1 className="font-serif text-2xl font-bold">Admin</h1>
			{seedData && (
				<p role="status" className="max-w-md rounded border border-border p-3 text-sm">
					Synthetic seed users are live in this database. Remove them with{" "}
					<code>SEED_TARGET=production bun lib/db/seed.ts --purge --i-know-this-is-production</code>, then
					turn the <code>seed-data</code> flag off.
				</p>
			)}
			<dl className="grid max-w-md grid-cols-[1fr_auto] text-sm">
				{rows.map(([label, value]) => (
					<div key={label} className="contents">
						<dt className="border-b border-border py-2">{label}</dt>
						<dd className="border-b border-border py-2 text-right font-mono tabular-nums">{value}</dd>
					</div>
				))}
			</dl>
			<p className="text-sm text-muted">
				Grant the admin role with <code>bun run admin:grant &lt;email&gt;</code>.
			</p>
		</section>
	);
}
