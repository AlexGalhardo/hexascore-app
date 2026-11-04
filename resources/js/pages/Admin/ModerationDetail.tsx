import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { ActionForm } from "../../components/admin/ActionForm";
import { Field } from "../../components/admin/Field";
import { PageHeader } from "../../components/admin/PageHeader";
import { Badge } from "../../components/admin/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/admin/ui/card";
import { Checkbox } from "../../components/admin/ui/checkbox";
import { Input } from "../../components/admin/ui/input";
import { Select } from "../../components/admin/ui/select";
import { Textarea } from "../../components/admin/ui/textarea";
import { useT } from "../../lib/useT";

interface Diff {
	field: string;
	kind: string;
	current: string;
	before: string;
	after: string;
	words: { op: string; text: string }[] | null;
	add: string[] | null;
	remove: string[] | null;
}

interface Detail {
	id: string;
	kind: string;
	targetType: string;
	targetID: string;
	domain: string;
	locale: string;
	status: string;
	note: string;
	rejectReason: string;
	decisionNote: string;
	sourceURLs: string[] | null;
	version: number;
	authorID: number;
	age: string;
	competing: boolean;
	stale: boolean;
	unreachable: boolean;
	diffs: Diff[] | null;
	reviews:
		| {
				reviewerId: number;
				recommendation: string;
				note: string;
				version: number;
				at: string;
		  }[]
		| null;
}

interface Props {
	detail: Detail;
	canApply: boolean;
	rejectReasons: string[] | null;
}

function Row({ label, children }: { label: string; children: React.ReactNode }): ReactElement {
	return (
		<div className="grid gap-0.5 sm:grid-cols-[10rem_1fr] sm:gap-3">
			<dt className="text-sm text-muted">{label}</dt>
			<dd className="text-sm">{children}</dd>
		</div>
	);
}

function DiffCard({ diff }: { diff: Diff }): ReactElement {
	const t = useT();
	// Words are keyed by their position in the text, which never changes for a stored diff.
	let pos = 0;
	const words = (diff.words ?? []).map((w) => {
		const item = { ...w, key: `${pos}:${w.op}` };
		pos += w.text.length + 1;
		return item;
	});
	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-base">{diff.field}</CardTitle>
			</CardHeader>
			<CardContent className="grid gap-2 text-sm">
				{diff.kind === "relation" ? (
					<p className="flex flex-wrap gap-1.5">
						{(diff.add ?? []).map((a) => (
							<Badge key={`+${a}`} variant="success">
								+ {a}
							</Badge>
						))}
						{(diff.remove ?? []).map((r) => (
							<Badge key={`-${r}`} variant="destructive">
								- {r}
							</Badge>
						))}
					</p>
				) : (
					<dl className="grid gap-1.5">
						<Row label={t("admin.suggestion.current")}>{diff.current || "–"}</Row>
						<Row label={t("admin.suggestion.before")}>{diff.before || "–"}</Row>
						<Row label={t("admin.suggestion.after")}>
							{diff.kind === "cover" || words.length === 0
								? diff.after || "–"
								: words.map((w) =>
										w.op === "del" ? (
											<del key={w.key} className="mr-1 text-danger">
												{w.text}
											</del>
										) : w.op === "add" ? (
											<ins
												key={w.key}
												className="mr-1 text-score-high no-underline"
											>
												{w.text}
											</ins>
										) : (
											<span key={w.key} className="mr-1">
												{w.text}
											</span>
										),
									)}
						</Row>
					</dl>
				)}
			</CardContent>
		</Card>
	);
}

export default function ModerationDetail({
	detail: d,
	canApply,
	rejectReasons,
}: Props): ReactElement {
	const t = useT();
	const base = `/admin/moderation/${d.id}`;
	const open = d.status === "submitted" || d.status === "in_review";
	const version = <input type="hidden" name="version" value={d.version} />;
	return (
		<>
			<Head title={t("admin.suggestion.title")} />
			<PageHeader
				title={t("admin.suggestion.title")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.moderation.title"), href: "/admin/moderation" },
					{ label: t("admin.suggestion.title") },
				]}
				actions={
					<span className="flex flex-wrap gap-1.5">
						{d.competing && (
							<Badge variant="warning">{t("admin.suggestion.competing")}</Badge>
						)}
						{d.stale && <Badge variant="warning">{t("admin.suggestion.stale")}</Badge>}
						{d.unreachable && (
							<Badge variant="destructive">{t("admin.suggestion.unreachable")}</Badge>
						)}
					</span>
				}
			/>
			<div className="grid gap-4">
				<Card>
					<CardContent>
						<dl className="grid gap-2">
							<Row label={t("admin.suggestion.status")}>
								{d.status} {d.age && <Badge variant="secondary">{d.age}</Badge>}
							</Row>
							<Row label={t("admin.suggestion.kind")}>{d.kind}</Row>
							<Row label={t("admin.suggestion.target")}>
								{d.targetType} {d.targetID} · {d.domain} · {d.locale}
							</Row>
							<Row label={t("admin.suggestion.author")}>
								{d.authorID ? `#${d.authorID}` : "–"}
							</Row>
							{d.note && <Row label={t("admin.suggestion.authorNote")}>{d.note}</Row>}
							{(d.sourceURLs ?? []).length > 0 && (
								<Row label={t("admin.suggestion.sources")}>
									<span className="flex flex-wrap gap-2">
										{(d.sourceURLs ?? []).map((u) => (
											<span key={u} className="break-all">
												{u}
											</span>
										))}
									</span>
								</Row>
							)}
							{d.rejectReason && (
								<Row label={t("admin.suggestion.rejected")}>
									{d.rejectReason}
									{d.decisionNote ? `: ${d.decisionNote}` : ""}
								</Row>
							)}
						</dl>
					</CardContent>
				</Card>

				<h2 className="text-lg font-semibold">{t("admin.suggestion.changes")}</h2>
				{(d.diffs ?? []).map((diff) => (
					<DiffCard key={diff.field} diff={diff} />
				))}

				<Card>
					<CardHeader>
						<CardTitle>{t("admin.suggestion.recommendations")}</CardTitle>
					</CardHeader>
					<CardContent>
						{(d.reviews ?? []).length === 0 ? (
							<p className="text-sm text-muted">{t("admin.suggestion.noRecs")}</p>
						) : (
							<ul className="grid gap-1.5 text-sm">
								{(d.reviews ?? []).map((r) => (
									<li key={`${r.reviewerId}-${r.at}`}>
										{t("admin.suggestion.recommends", {
											id: r.reviewerId,
											rec: r.recommendation,
											version: r.version,
											at: r.at.slice(0, 16).replace("T", " "),
										})}
										{r.note ? `: ${r.note}` : ""}
									</li>
								))}
							</ul>
						)}
					</CardContent>
				</Card>

				{open ? (
					<>
						<Card>
							<CardHeader>
								<CardTitle>{t("admin.suggestion.yourRec")}</CardTitle>
							</CardHeader>
							<CardContent>
								<ActionForm
									action={`${base}/recommend`}
									submit={t("admin.suggestion.recordRec")}
								>
									{version}
									<fieldset className="flex gap-4 text-sm">
										<legend className="sr-only">
											{t("admin.suggestion.yourRec")}
										</legend>
										<label className="flex items-center gap-2">
											<input
												type="radio"
												name="recommendation"
												value="approve"
												required
												className="accent-primary"
											/>
											{t("admin.suggestion.approve")}
										</label>
										<label className="flex items-center gap-2">
											<input
												type="radio"
												name="recommendation"
												value="reject"
												className="accent-primary"
											/>
											{t("admin.suggestion.reject")}
										</label>
									</fieldset>
									<Field label={t("admin.suggestion.note")}>
										{(id) => <Textarea id={id} name="note" maxLength={1000} />}
									</Field>
								</ActionForm>
							</CardContent>
						</Card>
						{canApply && (
							<div className="grid gap-4 lg:grid-cols-2">
								<Card>
									<CardHeader>
										<CardTitle>{t("admin.suggestion.approveBtn")}</CardTitle>
									</CardHeader>
									<CardContent>
										<ActionForm
											action={`${base}/approve`}
											submit={t("admin.suggestion.approveBtn")}
										>
											{version}
											<div className="flex items-center gap-2 text-sm">
												<Checkbox
													id="confirm-stale"
													name="confirm_stale"
													value="1"
												/>
												<label htmlFor="confirm-stale">
													{t("admin.suggestion.confirmStale")}
												</label>
											</div>
											{d.kind === "merge" && (
												<Field label={t("admin.suggestion.mergeConfirm")}>
													{(id) => (
														<Input
															id={id}
															name="confirm_title"
															required
															autoComplete="off"
														/>
													)}
												</Field>
											)}
										</ActionForm>
									</CardContent>
								</Card>
								<Card>
									<CardHeader>
										<CardTitle>{t("admin.suggestion.rejectBtn")}</CardTitle>
									</CardHeader>
									<CardContent>
										<ActionForm
											action={`${base}/reject`}
											submit={t("admin.suggestion.rejectBtn")}
											variant="destructive"
										>
											{version}
											<Field label={t("admin.suggestion.rejectReason")}>
												{(id) => (
													<Select
														id={id}
														name="reason"
														required
														defaultValue=""
													>
														<option value="">
															{t("admin.common.choose")}
														</option>
														{(rejectReasons ?? []).map((r) => (
															<option key={r} value={r}>
																{r}
															</option>
														))}
													</Select>
												)}
											</Field>
											<Field label={t("admin.suggestion.note")}>
												{(id) => (
													<Textarea
														id={id}
														name="note"
														maxLength={1000}
													/>
												)}
											</Field>
										</ActionForm>
									</CardContent>
								</Card>
							</div>
						)}
					</>
				) : (
					<ActionForm
						action={`${base}/clone`}
						submit={t("admin.suggestion.clone")}
						variant="outline"
					>
						<span className="hidden" />
					</ActionForm>
				)}
			</div>
		</>
	);
}
