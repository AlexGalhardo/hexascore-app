import { Head, usePage } from "@inertiajs/react";
import type { ReactElement } from "react";

import { ActionButton, ActionForm } from "../../components/admin/ActionForm";
import { Field } from "../../components/admin/Field";
import { PageHeader } from "../../components/admin/PageHeader";
import { CellValue, colLabel, type ResColumn } from "../../components/admin/resources";
import { buttonVariants } from "../../components/admin/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/admin/ui/card";
import { Input } from "../../components/admin/ui/input";
import { Select } from "../../components/admin/ui/select";
import { Textarea } from "../../components/admin/ui/textarea";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	resource: { slug: string; navKey: string };
	id: string;
	columns: ResColumn[];
	cells: unknown[];
}

const OUTCOMES = [
	"no_action",
	"dismissed",
	"list_hidden",
	"user_banned",
	"content_hidden",
	"content_deleted",
	"content_unhidden",
] as const;

const SECTIONS: Record<string, string> = {
	game: "games",
	movie: "movies",
	series: "series",
	book: "books",
};

/** The end of the chosen day in UTC as the RFC 3339 time the ban handler reads. */
function banExpiry(data: Record<string, string>): Record<string, string> {
	const { expires_day, ...rest } = data;
	return expires_day ? { ...rest, expires_at: `${expires_day}T23:59:59Z` } : rest;
}

function Actions({
	slug,
	id,
	values,
}: {
	slug: string;
	id: string;
	values: Record<string, unknown>;
}): ReactElement | null {
	const t = useT();
	const link = buttonVariants({ variant: "outline", size: "sm" });
	switch (slug) {
		case "users": {
			const banned = Boolean(values.banned_at);
			return (
				<Card>
					<CardHeader>
						<CardTitle>{t("admin.res.actions.banTitle")}</CardTitle>
					</CardHeader>
					<CardContent className="grid gap-4">
						{banned ? (
							<div>
								<ActionButton
									action={`/admin/users/${id}/unban`}
									variant="outline"
									size="default"
								>
									{t("admin.res.actions.unban")}
								</ActionButton>
							</div>
						) : (
							<ActionForm
								action={`/admin/users/${id}/ban`}
								submit={t("admin.res.actions.ban")}
								variant="destructive"
								transform={banExpiry}
								className="max-w-xl"
							>
								<Field label={t("admin.res.actions.banReason")}>
									{(fid) => <Input id={fid} name="reason" maxLength={500} />}
								</Field>
								<Field label={t("admin.res.actions.banUntil")}>
									{(fid) => <Input id={fid} name="expires_day" type="date" />}
								</Field>
							</ActionForm>
						)}
					</CardContent>
				</Card>
			);
		}
		case "reports":
			if (values.status !== "open") return null;
			return (
				<Card>
					<CardHeader>
						<CardTitle>{t("admin.res.actions.resolveTitle")}</CardTitle>
					</CardHeader>
					<CardContent>
						<ActionForm
							action={`/admin/reports/${id}/resolve`}
							submit={t("admin.res.actions.resolve")}
							className="max-w-xl"
						>
							<Field label={t("admin.res.actions.outcome")}>
								{(fid) => (
									<Select id={fid} name="outcome">
										{OUTCOMES.map((o) => (
											<option key={o} value={o}>
												{t(`admin.res.outcomes.${o}`)}
											</option>
										))}
									</Select>
								)}
							</Field>
							<Field label={t("admin.res.actions.adminNote")}>
								{(fid) => <Textarea id={fid} name="adminNote" maxLength={1000} />}
							</Field>
						</ActionForm>
					</CardContent>
				</Card>
			);
		case "score-bursts":
			if (values.resolution) return null;
			return (
				<Card>
					<CardHeader>
						<CardTitle>{t("admin.res.actions.burstTitle")}</CardTitle>
					</CardHeader>
					<CardContent className="flex gap-2">
						<ActionButton
							action={`/admin/scoring/bursts/${id}/approve`}
							variant="default"
							size="default"
						>
							{t("admin.res.actions.approve")}
						</ActionButton>
						<ActionButton
							action={`/admin/scoring/bursts/${id}/exclude`}
							variant="destructive"
							size="default"
						>
							{t("admin.res.actions.exclude")}
						</ActionButton>
					</CardContent>
				</Card>
			);
		case "items": {
			const section = SECTIONS[String(values.type)];
			return (
				<div className="flex flex-wrap gap-2">
					<a href={`/admin/items/${id}/manage`} className={link}>
						{t("admin.res.actions.manage")}
					</a>
					{section && (
						<a href={`/${section}/${String(values.slug)}`} className={link}>
							{t("admin.res.actions.viewOnSite")}
						</a>
					)}
				</div>
			);
		}
		case "contact-messages": {
			const handled = Boolean(values.handled);
			return (
				<div>
					<ActionButton
						action={`/admin/r/contact-messages/${id}/handled`}
						fields={{ handled: handled ? "false" : "true" }}
						variant={handled ? "outline" : "default"}
						size="default"
					>
						{handled
							? t("admin.res.actions.markUnhandled")
							: t("admin.res.actions.markHandled")}
					</ActionButton>
				</div>
			);
		}
		default:
			return null;
	}
}

export default function ResourceShow({ resource, id, columns, cells }: Props): ReactElement {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const title = t(`admin.nav.${resource.navKey}`);
	const values: Record<string, unknown> = Object.fromEntries(
		columns.map((c, i) => [c.key, cells[i]]),
	);
	const first = columns[0];
	const heading = first ? String(values[first.key] ?? id) : id;
	return (
		<>
			<Head title={`${title} ${heading}`} />
			<PageHeader
				title={`${title} · ${heading}`}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: title, href: `/admin/r/${resource.slug}` },
					{ label: heading },
				]}
			/>
			<div className="grid gap-4">
				<Actions slug={resource.slug} id={id} values={values} />
				<Card>
					<CardContent>
						<dl className="grid gap-3">
							{columns.map((c) => (
								<div
									key={c.key}
									className="grid gap-0.5 sm:grid-cols-[14rem_1fr] sm:gap-3"
								>
									<dt className="text-sm text-muted">{colLabel(t, c.key)}</dt>
									<dd className="min-w-0 text-sm break-words">
										<CellValue
											kind={c.kind}
											value={values[c.key]}
											locale={locale}
											t={t}
										/>
									</dd>
								</div>
							))}
						</dl>
					</CardContent>
				</Card>
			</div>
		</>
	);
}
