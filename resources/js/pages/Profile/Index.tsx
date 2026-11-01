import { Head, Link, router, useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import DataExport from "../../components/DataExport";
import DeleteAccount from "../../components/DeleteAccount";
import Field from "../../components/Field";
import PasswordField from "../../components/PasswordField";
import FeedOptIn from "../../components/social/FeedOptIn";
import { checkPassword, passwordAdvice } from "../../lib/password-policy";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import { createPasskey } from "../../lib/webauthn";

interface SessionRow {
	handle: string;
	current: boolean;
	lastSeenAt: string;
	userAgent: string;
}

interface PasskeyRow {
	id: number;
	name: string;
	createdAt: string;
	lastUsedAt: string | null;
}

interface Props {
	details: { name: string; email: string; locale: string };
	passwordSet: boolean;
	twoFactor: { enabled: boolean };
	passkeys?: PasskeyRow[];
	sessions: SessionRow[];
	privacy: { profilePrivate: boolean; personalizationOptOut: boolean };
	notifications: { communityMailOptOut: boolean };
	/** Present only while the feature flag `social` is on (SDD 14 FR-14-18). */
	social?: { feedOptIn: boolean };
}

const LOCALES = [
	["en", "English"],
	["pt-BR", "Português"],
	["es", "Español"],
] as const;

const button =
	"rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50";

function Section({
	id,
	title,
	children,
}: {
	id: string;
	title: string;
	children: React.ReactNode;
}) {
	return (
		<section id={id} aria-labelledby={`${id}-h`} className="mb-10 space-y-4">
			<h2 id={`${id}-h`} className="text-xl font-semibold">
				{title}
			</h2>
			{children}
		</section>
	);
}

function Passkeys({ passwordSet, passkeys }: { passwordSet: boolean; passkeys: PasskeyRow[] }) {
	const t = useT();
	const l = useLocalized();
	const [password, setPassword] = useState("");
	const [failed, setFailed] = useState(false);

	const add = async (): Promise<void> => {
		setFailed(false);

		try {
			const options = await fetch(l("/user/passkeys/options"), {
				method: "POST",
				credentials: "same-origin",
				headers: { Accept: "application/json", "Content-Type": "application/json" },
				body: JSON.stringify({ password }),
			});
			if (!options.ok) {
				setFailed(true);
				return;
			}

			const credential = await createPasskey(await options.json());
			const saved = await fetch(l("/user/passkeys"), {
				method: "POST",
				credentials: "same-origin",
				headers: { Accept: "application/json", "Content-Type": "application/json" },
				body: JSON.stringify(credential),
			});
			if (!saved.ok) {
				setFailed(true);
				return;
			}

			window.location.assign(saved.url);
		} catch {
			setFailed(true);
		}
	};

	return (
		<Section id="passkeys" title={t("account.passkeys.title")}>
			<p>{t("account.passkeys.body")}</p>
			{passkeys.length === 0 ? (
				<p>{t("account.passkeys.empty")}</p>
			) : (
				<ul className="divide-y divide-border">
					{passkeys.map((row) => (
						<li key={row.id} className="flex flex-wrap items-center gap-3 py-2">
							<span className="min-w-0 flex-1">
								{row.name || t("account.passkeys.unnamed")}
								<span className="ml-2 text-sm text-muted">
									{t("account.passkeys.added", {
										date: new Date(row.createdAt).toLocaleString(),
									})}
								</span>
							</span>
							<button
								type="button"
								className="font-medium underline"
								aria-label={t("account.passkeys.removeLabel", { name: row.name })}
								onClick={() => router.delete(l(`/profile/passkeys/${row.id}`))}
							>
								{t("account.passkeys.remove")}
							</button>
						</li>
					))}
				</ul>
			)}
			{passkeys.length === 1 && <p role="note">{t("account.passkeys.lastWarning")}</p>}
			{passwordSet ? (
				<PasswordField
					label={t("account.passkeys.password")}
					name="password"
					autoComplete="current-password"
					value={password}
					onChange={setPassword}
				/>
			) : (
				<Link href={l("/auth/google/confirm")} className="font-medium underline">
					{t("account.passkeys.confirmGoogle")}
				</Link>
			)}
			<button type="button" className={button} onClick={() => void add()}>
				{t("account.passkeys.add")}
			</button>
			{failed && (
				<p role="alert" className="mt-1 text-sm text-red-600">
					{t("account.passkeys.failed")}
				</p>
			)}
		</Section>
	);
}

export default function Index({
	details,
	passwordSet,
	twoFactor,
	passkeys = [],
	sessions,
	privacy,
	notifications,
	social,
}: Props) {
	const t = useT();
	const l = useLocalized();
	const detailsForm = useForm({ name: details.name, locale: details.locale });
	const passwordForm = useForm({ currentPassword: "", newPassword: "" });
	const privacyForm = useForm(privacy);
	const notificationsForm = useForm(notifications);

	const saveDetails = (e: FormEvent) => {
		e.preventDefault();
		detailsForm.patch(l("/profile/details"));
	};
	const savePassword = (e: FormEvent) => {
		e.preventDefault();
		passwordForm.put(l("/profile/password"), { onSuccess: () => passwordForm.reset() });
	};
	const savePrivacy = (e: FormEvent) => {
		e.preventDefault();
		privacyForm.patch(l("/profile/privacy"));
	};
	const saveNotifications = (e: FormEvent) => {
		e.preventDefault();
		notificationsForm.patch(l("/profile/notifications"));
	};

	return (
		<>
			<Head title={t("account.metaTitle")} />
			<h1 className="mb-6 text-2xl font-semibold">{t("account.title")}</h1>

			<Section id="details" title={t("account.details.title")}>
				<form onSubmit={saveDetails} className="space-y-4">
					<Field
						label={t("account.details.name")}
						name="name"
						autoComplete="name"
						required
						value={detailsForm.data.name}
						onChange={(e) => detailsForm.setData("name", e.target.value)}
						error={detailsForm.errors.name}
					/>
					<p>
						<span className="font-medium">{t("account.details.email")}</span>{" "}
						{details.email}
					</p>
					<label className="block">
						<span className="font-medium">{t("account.details.locale")}</span>
						<select
							value={detailsForm.data.locale}
							onChange={(e) => detailsForm.setData("locale", e.target.value)}
							className="mt-1 block rounded-md border border-zinc-300 bg-transparent px-3 py-2 dark:border-zinc-700"
						>
							{LOCALES.map(([code, label]) => (
								<option key={code} value={code}>
									{label}
								</option>
							))}
						</select>
					</label>
					<button type="submit" disabled={detailsForm.processing} className={button}>
						{t("account.details.save")}
					</button>
				</form>
			</Section>

			<Section
				id="password"
				title={
					passwordSet ? t("account.password.changeTitle") : t("account.password.setTitle")
				}
			>
				{!passwordSet && <p>{t("account.password.setBody")}</p>}
				<form onSubmit={savePassword} className="space-y-4">
					{passwordSet && (
						<PasswordField
							label={t("auth.fields.currentPassword")}
							name="currentPassword"
							autoComplete="current-password"
							value={passwordForm.data.currentPassword}
							onChange={(value) => passwordForm.setData("currentPassword", value)}
							error={passwordForm.errors.currentPassword}
						/>
					)}
					<PasswordField
						label={t("auth.fields.newPassword")}
						name="newPassword"
						autoComplete="new-password"
						value={passwordForm.data.newPassword}
						onChange={(value) => passwordForm.setData("newPassword", value)}
						error={passwordForm.errors.newPassword}
						checks={checkPassword(passwordForm.data.newPassword)}
						advice={passwordAdvice(passwordForm.data.newPassword)}
					/>
					<button type="submit" disabled={passwordForm.processing} className={button}>
						{passwordSet ? t("account.password.change") : t("account.password.set")}
					</button>
				</form>
			</Section>

			<Section id="two-factor" title={t("account.twoFactor.title")}>
				<p>{twoFactor.enabled ? t("account.twoFactor.on") : t("account.twoFactor.off")}</p>
				{passwordSet ? (
					<Link href="/user/two-factor" className="font-medium underline">
						{t("account.twoFactor.manage")}
					</Link>
				) : (
					<p role="note">{t("account.twoFactor.needPassword")}</p>
				)}
			</Section>

			<Passkeys passwordSet={passwordSet} passkeys={passkeys} />

			<Section id="sessions" title={t("account.sessions.title")}>
				{sessions.length === 0 ? (
					<p>{t("account.sessions.empty")}</p>
				) : (
					<ul className="divide-y divide-border">
						{sessions.map((s) => (
							<li key={s.handle} className="flex flex-wrap items-center gap-3 py-2">
								<span className="min-w-0 flex-1 truncate">
									{s.current
										? t("account.sessions.current")
										: s.userAgent || s.handle}
									<span className="ml-2 text-sm text-muted">
										{t("account.sessions.lastSeen")}{" "}
										{new Date(s.lastSeenAt).toLocaleString()}
									</span>
								</span>
								{!s.current && (
									<button
										type="button"
										className="font-medium underline"
										onClick={() =>
											router.delete(l(`/profile/sessions/${s.handle}`))
										}
									>
										{t("account.sessions.revoke")}
									</button>
								)}
							</li>
						))}
					</ul>
				)}
				<button
					type="button"
					className={button}
					onClick={() => router.post(l("/profile/sessions/revoke-all"))}
				>
					{t("account.sessions.revokeAll")}
				</button>
			</Section>

			<Section id="privacy" title={t("account.privacy.title")}>
				<form onSubmit={savePrivacy} className="space-y-4">
					<label className="flex items-start gap-2">
						<input
							type="checkbox"
							checked={privacyForm.data.profilePrivate}
							onChange={(e) =>
								privacyForm.setData("profilePrivate", e.target.checked)
							}
						/>
						<span>
							<span className="font-medium">{t("account.privacy.label")}</span>
							<span className="block text-sm text-muted">
								{t("account.privacy.help")}
							</span>
						</span>
					</label>
					<label className="flex items-start gap-2">
						<input
							type="checkbox"
							checked={privacyForm.data.personalizationOptOut}
							onChange={(e) =>
								privacyForm.setData("personalizationOptOut", e.target.checked)
							}
						/>
						<span>
							<span className="font-medium">
								{t("account.privacy.personalization")}
							</span>
							<span className="block text-sm text-muted">
								{t("account.privacy.personalizationHelp")}
							</span>
						</span>
					</label>
					<button type="submit" disabled={privacyForm.processing} className={button}>
						{t("account.privacy.save")}
					</button>
				</form>
				{social && (
					<div className="mt-6 border-t border-border pt-4">
						<FeedOptIn on={social.feedOptIn} />
					</div>
				)}
			</Section>

			<Section id="notifications" title={t("account.notifications.title")}>
				<form onSubmit={saveNotifications} className="space-y-4">
					<label className="flex items-start gap-2">
						<input
							type="checkbox"
							checked={notificationsForm.data.communityMailOptOut}
							onChange={(e) =>
								notificationsForm.setData("communityMailOptOut", e.target.checked)
							}
						/>
						<span>
							<span className="font-medium">
								{t("account.notifications.community")}
							</span>
							<span className="block text-sm text-muted">
								{t("account.notifications.communityHelp")}
							</span>
						</span>
					</label>
					<button
						type="submit"
						disabled={notificationsForm.processing}
						className={button}
					>
						{t("account.notifications.save")}
					</button>
				</form>
			</Section>

			<Section id="export" title={t("account.export.title")}>
				<DataExport className={button} />
			</Section>

			<Section id="delete" title={t("account.delete.title")}>
				<DeleteAccount className={button} />
			</Section>
		</>
	);
}
