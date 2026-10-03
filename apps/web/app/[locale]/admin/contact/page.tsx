import { setContactHandled } from "@/actions/admin";
import { ActionForm } from "@/components/admin/action-form";
import { Link } from "@/i18n/navigation";
import { listContactMessages } from "@/lib/admin";
import { db } from "@/lib/db/client";
import { button, cell, head, requireAdmin } from "../guard";

type Props = { searchParams: Promise<{ handled?: string }> };

export default async function AdminContactPage({ searchParams }: Props) {
	await requireAdmin();
	const handled = (await searchParams).handled === "1";
	const messages = await listContactMessages(db, handled);
	return (
		<section className="grid gap-4">
			<div className="flex flex-wrap items-baseline gap-4">
				<h1 className="font-serif text-2xl font-bold">Contact messages</h1>
				<Link href={handled ? "/admin/contact" : "/admin/contact?handled=1"} className="text-sm underline">
					{handled ? "Show unhandled" : "Show handled"}
				</Link>
			</div>
			<div className="overflow-x-auto">
				<table className="w-full min-w-[40rem] text-sm">
					<thead>
						<tr>
							<th className={head}>Date</th>
							<th className={head}>From</th>
							<th className={head}>Subject</th>
							<th className={head}>Message</th>
							<th className={head} />
						</tr>
					</thead>
					<tbody>
						{messages.map((m) => (
							<tr key={m.id}>
								<td className={`${cell} whitespace-nowrap font-mono text-xs`}>
									{m.createdAt.slice(0, 16).replace("T", " ")}
								</td>
								<td className={cell}>
									{m.name}
									<br />
									<a href={`mailto:${m.email}`} className="text-xs underline">
										{m.email}
									</a>
								</td>
								<td className={cell}>{m.subject}</td>
								<td className={`${cell} max-w-md whitespace-pre-line`}>{m.message}</td>
								<td className={cell}>
									<ActionForm action={setContactHandled}>
										<input type="hidden" name="id" value={m.id} />
										<input type="hidden" name="handled" value={handled ? "false" : "true"} />
										<button type="submit" className={button}>
											{handled ? "Reopen" : "Mark handled"}
										</button>
									</ActionForm>
								</td>
							</tr>
						))}
					</tbody>
				</table>
				{messages.length ? null : <p className="py-4 text-muted">Nothing here.</p>}
			</div>
		</section>
	);
}
