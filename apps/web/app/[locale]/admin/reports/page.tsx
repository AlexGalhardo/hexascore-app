import { banUser, setListHidden, setReportStatus } from "@/actions/admin";
import { ActionForm } from "@/components/admin/action-form";
import { Link } from "@/i18n/navigation";
import { listReports } from "@/lib/admin";
import { db } from "@/lib/db/client";
import { button, cell, head, input, requireAdmin } from "../guard";

type Props = { searchParams: Promise<{ status?: string }> };

export default async function AdminReportsPage({ searchParams }: Props) {
	await requireAdmin();
	const closed = (await searchParams).status === "closed";
	const reports = await listReports(db, closed ? "closed" : "open");
	return (
		<section className="grid gap-4">
			<div className="flex flex-wrap items-baseline gap-4">
				<h1 className="font-serif text-2xl font-bold">{closed ? "Closed reports" : "Open reports"}</h1>
				<Link href={closed ? "/admin/reports" : "/admin/reports?status=closed"} className="text-sm underline">
					{closed ? "Show open" : "Show closed"}
				</Link>
			</div>
			<div className="overflow-x-auto">
				<table className="w-full min-w-[48rem] text-sm">
					<thead>
						<tr>
							<th className={head}>Date</th>
							<th className={head}>Target</th>
							<th className={head}>Reason</th>
							<th className={head}>Reporter</th>
							<th className={head}>Actions</th>
						</tr>
					</thead>
					<tbody>
						{reports.map((r) => (
							<tr key={r.id}>
								<td className={`${cell} whitespace-nowrap font-mono text-xs`}>
									{r.createdAt.slice(0, 16).replace("T", " ")}
								</td>
								<td className={cell}>
									<span className="text-xs uppercase text-muted">{r.targetType}</span>{" "}
									{r.targetLabel ?? <em className="text-muted">deleted</em>}
									{r.listHidden ? <span className="ml-1 text-xs text-danger">hidden</span> : null}
									<br />
									<code className="text-xs text-muted">{r.targetId}</code>
								</td>
								<td className={cell}>
									{r.reason}
									{r.details ? (
										<p className="max-w-xs whitespace-pre-line text-xs">{r.details}</p>
									) : null}
								</td>
								<td className={cell}>{r.reporter ?? "—"}</td>
								<td className={`${cell} grid gap-1`}>
									{r.status === "open" ? (
										<div className="flex flex-wrap gap-1">
											<ActionForm action={setReportStatus}>
												<input type="hidden" name="id" value={r.id} />
												<input type="hidden" name="status" value="resolved" />
												<button type="submit" className={button}>
													Resolve
												</button>
											</ActionForm>
											<ActionForm action={setReportStatus}>
												<input type="hidden" name="id" value={r.id} />
												<input type="hidden" name="status" value="dismissed" />
												<button type="submit" className={button}>
													Dismiss
												</button>
											</ActionForm>
										</div>
									) : (
										<span className="text-xs text-muted">{r.status}</span>
									)}
									{r.targetType === "list" && r.listHidden !== null ? (
										<ActionForm action={setListHidden}>
											<input type="hidden" name="listId" value={r.targetId} />
											<input
												type="hidden"
												name="hidden"
												value={r.listHidden ? "false" : "true"}
											/>
											<button type="submit" className={button}>
												{r.listHidden ? "Unhide list" : "Hide list"}
											</button>
										</ActionForm>
									) : null}
									{r.targetOwnerId ? (
										<ActionForm
											action={banUser}
											confirm="Ban this user? All their sessions are revoked."
										>
											<input type="hidden" name="userId" value={r.targetOwnerId} />
											<input
												name="reason"
												required
												maxLength={200}
												placeholder="Ban reason"
												aria-label="Ban reason"
												className={input}
											/>
											<button type="submit" className={`${button} text-danger`}>
												Ban {r.targetType === "list" ? "owner" : "user"}
											</button>
										</ActionForm>
									) : null}
								</td>
							</tr>
						))}
					</tbody>
				</table>
				{reports.length ? null : <p className="py-4 text-muted">Nothing here.</p>}
			</div>
		</section>
	);
}
