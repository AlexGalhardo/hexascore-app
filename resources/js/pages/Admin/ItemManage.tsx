import { Head } from "@inertiajs/react";
import type { ReactElement, ReactNode } from "react";

import { ActionButton, ActionForm } from "../../components/admin/ActionForm";
import { Field } from "../../components/admin/Field";
import { PageHeader } from "../../components/admin/PageHeader";
import { Badge } from "../../components/admin/ui/badge";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../../components/admin/ui/card";
import { Checkbox } from "../../components/admin/ui/checkbox";
import { Input } from "../../components/admin/ui/input";
import { Select } from "../../components/admin/ui/select";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "../../components/admin/ui/table";
import { useT } from "../../lib/useT";

interface Props {
	item: { id: string; type: string; slug: string };
	affiliates: { store: string; url: string; active: boolean }[] | null;
	trending: { id: number; kind: string; until: string; reason: string }[] | null;
	offers:
		| {
				id: string;
				store: string;
				storeName: string;
				country: string;
				kind: string;
				source: string;
				url: string;
				status: string;
				active: boolean;
		  }[]
		| null;
	stores: { code: string; name: string }[] | null;
	offerKinds: string[] | null;
	offerCountries: string[] | null;
}

const FIELDS = [
	["title", "title"],
	["release_date", "releaseDate"],
	["translations.en.title", "titleEn"],
	["translations.pt-BR.title", "titlePt"],
	["translations.es.title", "titleEs"],
	["external_ids.wikidata", "wikidata"],
] as const;

const AFFILIATE_STORES = ["amazon", "mercadolivre", "steam", "gog"];

function Section({
	title,
	hint,
	children,
	danger = false,
}: {
	title: string;
	hint?: string;
	children: ReactNode;
	danger?: boolean;
}): ReactElement {
	return (
		<Card className={danger ? "border-danger/50" : undefined}>
			<CardHeader>
				<CardTitle className={danger ? "text-danger" : undefined}>{title}</CardTitle>
				{hint && <CardDescription>{hint}</CardDescription>}
			</CardHeader>
			<CardContent className="grid gap-4">{children}</CardContent>
		</Card>
	);
}

export default function ItemManage({
	item,
	affiliates,
	trending,
	offers,
	stores,
	offerKinds,
	offerCountries,
}: Props): ReactElement {
	const t = useT();
	const base = `/admin/items/${item.id}`;
	const yesNo = (v: boolean) => (v ? t("admin.common.yes") : t("admin.common.no"));
	return (
		<>
			<Head title={t("admin.item.title")} />
			<PageHeader
				title={t("admin.item.title")}
				description={`${item.type} · ${item.slug}`}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.nav.items"), href: "/admin/items" },
					{ label: t("admin.item.title") },
				]}
				actions={
					<a
						href={base}
						className="text-sm text-primary underline-offset-4 hover:underline"
					>
						{t("admin.item.backToItem")}
					</a>
				}
			/>
			<div className="grid gap-4">
				<Section title={t("admin.item.editField")} hint={t("admin.item.editHint")}>
					<ActionForm
						action={`/admin/catalog/item/${item.id}/fields`}
						submit={t("admin.item.saveField")}
						className="md:grid-cols-2"
					>
						<Field label={t("admin.item.field")}>
							{(id) => (
								<Select id={id} name="field">
									{FIELDS.map(([value, key]) => (
										<option key={value} value={value}>
											{t(`admin.item.fields.${key}`)}
										</option>
									))}
								</Select>
							)}
						</Field>
						<Field label={t("admin.item.value")}>
							{(id) => <Input id={id} name="value" maxLength={500} />}
						</Field>
					</ActionForm>
				</Section>

				<Section title={t("admin.item.affiliates")}>
					{(affiliates ?? []).length === 0 ? (
						<p className="text-sm text-muted">{t("admin.item.noAffiliates")}</p>
					) : (
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>{t("admin.item.store")}</TableHead>
									<TableHead>{t("admin.item.url")}</TableHead>
									<TableHead>{t("admin.item.active")}</TableHead>
									<TableHead className="text-right">
										{t("admin.common.actions")}
									</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{(affiliates ?? []).map((a) => (
									<TableRow key={a.store}>
										<TableCell className="font-medium">{a.store}</TableCell>
										<TableCell className="max-w-72 truncate text-muted">
											{a.url}
										</TableCell>
										<TableCell>{yesNo(a.active)}</TableCell>
										<TableCell className="text-right">
											<ActionButton
												action={`${base}/affiliate-links/${a.store}/delete`}
											>
												{t("admin.common.delete")}
											</ActionButton>
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					)}
					<ActionForm
						action={`${base}/affiliate-links`}
						submit={t("admin.item.saveLink")}
						className="md:grid-cols-2"
					>
						<Field label={t("admin.item.store")}>
							{(id) => (
								<Select id={id} name="store">
									{AFFILIATE_STORES.map((s) => (
										<option key={s}>{s}</option>
									))}
								</Select>
							)}
						</Field>
						<Field label={t("admin.item.url")}>
							{(id) => (
								<Input
									id={id}
									name="url"
									type="url"
									required
									maxLength={2048}
									placeholder="https://"
								/>
							)}
						</Field>
						<div className="flex items-center gap-2 text-sm">
							<Checkbox
								id="affiliate-active"
								name="active"
								value="true"
								defaultChecked
							/>
							<label htmlFor="affiliate-active">{t("admin.item.active")}</label>
						</div>
					</ActionForm>
				</Section>

				<Section title={t("admin.item.offers")} hint={t("admin.item.offersHint")}>
					{(offers ?? []).length === 0 ? (
						<p className="text-sm text-muted">{t("admin.item.noOffers")}</p>
					) : (
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>{t("admin.item.store")}</TableHead>
									<TableHead>{t("admin.item.country")}</TableHead>
									<TableHead>{t("admin.item.kind")}</TableHead>
									<TableHead>{t("admin.item.sourceCol")}</TableHead>
									<TableHead>{t("admin.item.check")}</TableHead>
									<TableHead>{t("admin.item.url")}</TableHead>
									<TableHead>{t("admin.item.active")}</TableHead>
									<TableHead className="text-right">
										{t("admin.common.actions")}
									</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{(offers ?? []).map((o) => (
									<TableRow key={o.id}>
										<TableCell className="font-medium">{o.storeName}</TableCell>
										<TableCell>{o.country}</TableCell>
										<TableCell>{o.kind}</TableCell>
										<TableCell className="text-muted">{o.source}</TableCell>
										<TableCell>
											<Badge variant="secondary">{o.status}</Badge>
										</TableCell>
										<TableCell className="max-w-56 truncate text-muted">
											{o.url}
										</TableCell>
										<TableCell>{yesNo(o.active)}</TableCell>
										<TableCell className="text-right">
											<ActionButton action={`${base}/offers/${o.id}/toggle`}>
												{o.active
													? t("admin.item.deactivate")
													: t("admin.item.restore")}
											</ActionButton>
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					)}
					<ActionForm
						action={`${base}/offers`}
						submit={t("admin.item.saveOffer")}
						className="md:grid-cols-2"
					>
						<Field label={t("admin.item.store")}>
							{(id) => (
								<Select id={id} name="store">
									{(stores ?? []).map((s) => (
										<option key={s.code} value={s.code}>
											{s.name}
										</option>
									))}
								</Select>
							)}
						</Field>
						<Field label={t("admin.item.country")}>
							{(id) => (
								<Select id={id} name="country">
									{(offerCountries ?? []).map((c) => (
										<option key={c}>{c}</option>
									))}
								</Select>
							)}
						</Field>
						<Field label={t("admin.item.kind")}>
							{(id) => (
								<Select id={id} name="kind">
									{(offerKinds ?? []).map((k) => (
										<option key={k}>{k}</option>
									))}
								</Select>
							)}
						</Field>
						<Field label={t("admin.item.url")}>
							{(id) => (
								<Input
									id={id}
									name="url"
									type="url"
									required
									maxLength={2048}
									placeholder="https://"
								/>
							)}
						</Field>
					</ActionForm>
				</Section>

				<Section title={t("admin.item.trending")} hint={t("admin.item.trendingHint")}>
					{(trending ?? []).length === 0 ? (
						<p className="text-sm text-muted">{t("admin.item.noTrending")}</p>
					) : (
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>{t("admin.item.kind")}</TableHead>
									<TableHead>{t("admin.item.until")}</TableHead>
									<TableHead>{t("admin.common.reason")}</TableHead>
									<TableHead className="text-right">
										{t("admin.common.actions")}
									</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{(trending ?? []).map((o) => (
									<TableRow key={o.id}>
										<TableCell className="font-medium">{o.kind}</TableCell>
										<TableCell className="text-muted">
											{o.until || t("admin.item.noEnd")}
										</TableCell>
										<TableCell className="text-muted">{o.reason}</TableCell>
										<TableCell className="text-right">
											<ActionButton
												action={`${base}/trending/${o.id}/delete`}
											>
												{t("admin.common.remove")}
											</ActionButton>
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					)}
					<ActionForm
						action={`${base}/trending`}
						submit={t("admin.item.save")}
						className="md:grid-cols-3"
					>
						<Field label={t("admin.item.kind")}>
							{(id) => (
								<Select id={id} name="kind">
									<option value="pinned">{t("admin.item.pin")}</option>
									<option value="banned">{t("admin.item.ban")}</option>
								</Select>
							)}
						</Field>
						<Field label={t("admin.item.untilField")}>
							{(id) => (
								<Input
									id={id}
									name="until"
									maxLength={10}
									placeholder="2026-12-31"
								/>
							)}
						</Field>
						<Field label={t("admin.common.reason")}>
							{(id) => <Input id={id} name="reason" maxLength={200} />}
						</Field>
					</ActionForm>
				</Section>

				<Section title={t("admin.item.mergeTitle")} hint={t("admin.item.mergeHint")} danger>
					<ActionForm
						action={`${base}/merge`}
						submit={t("admin.item.mergeAction")}
						variant="destructive"
					>
						<Field label={t("admin.item.mergeTarget")} className="max-w-xl">
							{(id) => (
								<Input
									id={id}
									name="into"
									required
									pattern="[0-9a-fA-F\-]{36}"
									maxLength={36}
								/>
							)}
						</Field>
						<div className="flex items-center gap-2 text-sm">
							<Checkbox id="merge-confirm" name="confirm" value="merge" required />
							<label htmlFor="merge-confirm">{t("admin.item.mergeConfirm")}</label>
						</div>
					</ActionForm>
				</Section>
			</div>
		</>
	);
}
