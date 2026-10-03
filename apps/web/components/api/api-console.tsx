"use client";

import { useLocale, useTranslations } from "next-intl";
import { type FormEvent, type ReactNode, useRef, useState, useTransition } from "react";
import { createApiKeyAction, revokeApiKeyAction } from "@/actions/api-keys";
import { useRouter } from "@/i18n/navigation";
import type { ApiKeyView } from "@/lib/api/dashboard";
import { ApiReference } from "./api-reference";
import { CopyButton } from "./copy-button";

const KEY_PATTERN = /^hx_live_[0-9A-Za-z]{32}$/;

/**
 * Keys + "Try it". The full key exists only in this component's state: shown once after creation and handed to
 * Scalar in memory; nothing is written to storage or sent anywhere but the API itself.
 */
export function ApiConsole({ keys, children }: { keys: ApiKeyView[]; children: ReactNode }) {
	const t = useTranslations("api.profile");
	const td = useTranslations("api.docs");
	const locale = useLocale();
	const router = useRouter();
	const [pending, start] = useTransition();
	const [error, setError] = useState<string | null>(null);
	const [created, setCreated] = useState<string | null>(null);
	const [tryKey, setTryKey] = useState("");
	const [appliedKey, setAppliedKey] = useState<string | undefined>(undefined);
	const [revoking, setRevoking] = useState<ApiKeyView | null>(null);
	const dialog = useRef<HTMLDialogElement>(null);

	const date = (iso: string | null) =>
		iso ? new Intl.DateTimeFormat(locale, { dateStyle: "medium", timeStyle: "short" }).format(new Date(iso)) : null;

	function create(event: FormEvent<HTMLFormElement>): void {
		event.preventDefault();
		const form = event.currentTarget;
		const name = String(new FormData(form).get("name") ?? "");
		setError(null);
		start(async () => {
			const result = await createApiKeyAction(name);
			if (!result.ok) {
				setError(t(`keys.errors.${result.error}`));
				return;
			}
			form.reset();
			setCreated(result.key);
			setTryKey(result.key);
			router.refresh();
		});
	}

	function askRevoke(key: ApiKeyView): void {
		setRevoking(key);
		dialog.current?.showModal();
	}

	function confirmRevoke(): void {
		const key = revoking;
		if (!key) return;
		setError(null);
		start(async () => {
			const result = await revokeApiKeyAction(key.id);
			dialog.current?.close();
			setRevoking(null);
			if (!result.ok) setError(t(`keys.errors.${result.error}`));
			router.refresh();
		});
	}

	function applyTryKey(event: FormEvent<HTMLFormElement>): void {
		event.preventDefault();
		setAppliedKey(KEY_PATTERN.test(tryKey.trim()) ? tryKey.trim() : undefined);
	}

	return (
		<div className="space-y-10">
			<section aria-labelledby="keys-title" className="space-y-4">
				<h2 id="keys-title" className="font-serif text-2xl font-bold">
					{t("keys.title")}
				</h2>

				{created && (
					<div role="status" className="space-y-3 rounded-lg border-2 border-score-mid p-4">
						<p className="font-semibold">{t("keys.createdTitle")}</p>
						<div className="flex flex-wrap items-center gap-2">
							<code className="break-all rounded bg-surface px-2 py-1 font-mono text-sm select-all">
								{created}
							</code>
							<CopyButton text={created} label={t("keys.copy")} copiedLabel={t("keys.copied")} />
						</div>
						<p className="text-sm">{t("keys.createdWarning")}</p>
						<button
							type="button"
							onClick={() => setCreated(null)}
							className="min-h-11 rounded-md border border-border px-4 text-sm hover:bg-surface"
						>
							{t("keys.dismiss")}
						</button>
					</div>
				)}

				{keys.length === 0 ? (
					<p className="text-muted">{t("keys.empty")}</p>
				) : (
					<div className="overflow-x-auto">
						<table className="w-full border-collapse text-left text-sm">
							<thead>
								<tr className="border-b border-border">
									<th scope="col" className="py-2 pr-4">
										{t("keys.name")}
									</th>
									<th scope="col" className="py-2 pr-4">
										{t("keys.prefix")}
									</th>
									<th scope="col" className="py-2 pr-4">
										{t("keys.created")}
									</th>
									<th scope="col" className="py-2 pr-4">
										{t("keys.lastUsed")}
									</th>
									<th scope="col" className="py-2 pr-4">
										{t("keys.status")}
									</th>
									<th scope="col" className="py-2">
										<span className="sr-only">{t("keys.actions")}</span>
									</th>
								</tr>
							</thead>
							<tbody>
								{keys.map((key) => (
									<tr
										key={key.id}
										className={`border-b border-border ${key.revokedAt ? "text-muted" : ""}`}
									>
										<td className="py-2 pr-4 font-medium">{key.name}</td>
										<td className="py-2 pr-4 font-mono">{key.prefix}…</td>
										<td className="py-2 pr-4">{date(key.createdAt)}</td>
										<td className="py-2 pr-4">{date(key.lastUsedAt) ?? t("keys.never")}</td>
										<td className="py-2 pr-4">
											{key.revokedAt ? t("keys.revoked") : t("keys.active")}
										</td>
										<td className="py-2 text-right">
											{!key.revokedAt && (
												<button
													type="button"
													onClick={() => askRevoke(key)}
													disabled={pending}
													className="min-h-11 rounded-md px-3 text-danger hover:bg-surface disabled:opacity-50"
												>
													{t("keys.revoke")}
												</button>
											)}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				)}

				<form onSubmit={create} className="flex flex-wrap items-end gap-2">
					<label className="grid gap-1 text-sm">
						{t("keys.newName")}
						<input
							name="name"
							required
							maxLength={40}
							placeholder={t("keys.newNamePlaceholder")}
							className="min-h-11 rounded-md border border-border bg-bg px-3"
						/>
					</label>
					<button
						type="submit"
						disabled={pending}
						className="min-h-11 rounded-md bg-fg px-4 text-sm font-semibold text-bg disabled:opacity-50"
					>
						{pending ? t("keys.creating") : t("keys.create")}
					</button>
				</form>
				{error && (
					<p role="alert" className="text-sm text-danger">
						{error}
					</p>
				)}
			</section>

			{children}

			<section aria-labelledby="try-title" className="space-y-4">
				<h2 id="try-title" className="font-serif text-2xl font-bold">
					{t("tryIt.title")}
				</h2>
				<p className="max-w-prose text-sm text-muted">{t("tryIt.description")}</p>
				<form onSubmit={applyTryKey} className="flex flex-wrap items-end gap-2">
					<label className="grid gap-1 text-sm">
						{t("tryIt.keyLabel")}
						<input
							type="password"
							autoComplete="off"
							spellCheck={false}
							value={tryKey}
							onChange={(e) => setTryKey(e.target.value)}
							placeholder="hx_live_…"
							aria-describedby="try-key-hint"
							className="min-h-11 w-80 max-w-full rounded-md border border-border bg-bg px-3 font-mono"
						/>
					</label>
					<button
						type="submit"
						className="min-h-11 rounded-md border border-border px-4 text-sm hover:bg-surface"
					>
						{t("tryIt.apply")}
					</button>
				</form>
				<p id="try-key-hint" className="text-xs text-muted">
					{t("tryIt.keyHint")}
				</p>
				{/* Mounted only on demand: Scalar scrolls the window to its own hash routes when it loads. */}
				{appliedKey && (
					<div className="rounded-lg border border-border">
						<ApiReference loadingLabel={td("loading")} token={appliedKey} serverUrl="/api/v1" />
					</div>
				)}
			</section>

			<dialog
				ref={dialog}
				onClose={() => setRevoking(null)}
				aria-labelledby="revoke-title"
				className="m-auto w-[min(28rem,calc(100vw-2rem))] rounded-lg border border-border bg-bg p-6 text-fg backdrop:bg-black/50"
			>
				<h2 id="revoke-title" className="font-serif text-xl font-bold">
					{t("keys.confirmTitle", { name: revoking?.name ?? "" })}
				</h2>
				<p className="mt-2 text-sm">{t("keys.confirmBody")}</p>
				<div className="mt-6 flex flex-wrap justify-end gap-2">
					<button
						type="button"
						onClick={() => dialog.current?.close()}
						className="min-h-11 rounded-md border border-border px-4 text-sm hover:bg-surface"
					>
						{t("keys.cancel")}
					</button>
					<button
						type="button"
						onClick={confirmRevoke}
						disabled={pending}
						className="min-h-11 rounded-md bg-danger px-4 text-sm font-semibold text-bg disabled:opacity-50"
					>
						{pending ? t("keys.revoking") : t("keys.revoke")}
					</button>
				</div>
			</dialog>
		</div>
	);
}
