import { Head, Link, router, useForm, usePage } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import Field from "../../components/Field";
import { useT } from "../../lib/useT";

interface ApiKey {
	id: string;
	name: string;
	prefix: string;
	lastUsedAt: string | null;
	revokedAt: string | null;
	createdAt: string;
}

interface Props {
	keys: ApiKey[];
	maxKeys: number;
}

const button =
	"rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50";

// Server codes on errors.name that have their own message; anything else is a length error.
const NAME_ERRORS = new Set(["limit"]);

export default function ApiKeys({ keys, maxKeys }: Props) {
	const t = useT();
	const flash = usePage().flash as { newKey?: string } | undefined;
	const form = useForm({ name: "" });
	const [copied, setCopied] = useState(false);
	const [dismissed, setDismissed] = useState<string | null>(null);
	const [confirming, setConfirming] = useState<string | null>(null);
	const newKey = flash?.newKey && flash.newKey !== dismissed ? flash.newKey : null;
	const live = keys.filter((k) => !k.revokedAt).length;
	const nameError = form.errors.name;

	const create = (e: FormEvent) => {
		e.preventDefault();
		setCopied(false);
		form.post("/profile/api/keys", { onSuccess: () => form.reset() });
	};

	const copy = async (value: string) => {
		await navigator.clipboard.writeText(value);
		setCopied(true);
	};

	const date = (iso: string) => new Date(iso).toLocaleDateString();

	return (
		<>
			<Head title={t("api.profile.metaTitle")} />
			<h1 className="mb-2 text-2xl font-semibold">{t("api.profile.title")}</h1>
			<p className="mb-6">
				{t("api.profile.intro")}{" "}
				<Link href="/docs/api" className="underline">
					{t("api.profile.docsLink")}
				</Link>
			</p>

			{newKey && (
				<section
					aria-labelledby="new-key-h"
					className="mb-8 space-y-3 rounded-md border border-amber-500 p-4"
				>
					<h2 id="new-key-h" className="text-lg font-semibold">
						{t("api.profile.keys.createdTitle")}
					</h2>
					<p role="alert">{t("api.profile.keys.createdWarning")}</p>
					<code className="block break-all rounded bg-black/5 p-2 dark:bg-white/10">
						{newKey}
					</code>
					<div className="flex gap-3">
						<button type="button" className={button} onClick={() => copy(newKey)}>
							{copied ? t("api.profile.keys.copied") : t("api.profile.keys.copy")}
						</button>
						<button
							type="button"
							className="underline"
							onClick={() => setDismissed(newKey)}
						>
							{t("api.profile.keys.dismiss")}
						</button>
					</div>
				</section>
			)}

			<section aria-labelledby="keys-h" className="space-y-4">
				<h2 id="keys-h" className="text-xl font-semibold">
					{t("api.profile.keys.title")}
				</h2>

				<form onSubmit={create} className="flex flex-wrap items-end gap-3">
					<Field
						label={t("api.profile.keys.newName")}
						name="name"
						value={form.data.name}
						onChange={(e) => form.setData("name", e.target.value)}
						placeholder={t("api.profile.keys.newNamePlaceholder")}
						maxLength={40}
						required
						error={
							nameError &&
							t(
								`api.profile.keys.errors.${NAME_ERRORS.has(nameError) ? nameError : "invalid"}`,
								{
									count: maxKeys,
								},
							)
						}
					/>
					<button
						type="submit"
						className={button}
						disabled={form.processing || live >= maxKeys}
					>
						{form.processing
							? t("api.profile.keys.creating")
							: t("api.profile.keys.create")}
					</button>
				</form>

				{keys.length === 0 ? (
					<p>{t("api.profile.keys.empty")}</p>
				) : (
					<section
						className="overflow-x-auto"
						aria-labelledby="keys-h"
						// biome-ignore lint/a11y/noNoninteractiveTabindex: a scrollable region must be keyboard focusable (WCAG 2.1.1)
						tabIndex={0}
					>
						<table className="w-full text-left">
							<thead>
								<tr>
									<th scope="col">{t("api.profile.keys.name")}</th>
									<th scope="col">{t("api.profile.keys.prefix")}</th>
									<th scope="col">{t("api.profile.keys.created")}</th>
									<th scope="col">{t("api.profile.keys.lastUsed")}</th>
									<th scope="col">{t("api.profile.keys.status")}</th>
									<th scope="col">{t("api.profile.keys.actions")}</th>
								</tr>
							</thead>
							<tbody>
								{keys.map((k) => (
									<tr key={k.id}>
										<td>{k.name}</td>
										<td>
											<code>{k.prefix}…</code>
										</td>
										<td>{date(k.createdAt)}</td>
										<td>
											{k.lastUsedAt
												? date(k.lastUsedAt)
												: t("api.profile.keys.never")}
										</td>
										<td>
											{k.revokedAt
												? t("api.profile.keys.revoked")
												: t("api.profile.keys.active")}
										</td>
										<td>
											{!k.revokedAt &&
												(confirming === k.id ? (
													<fieldset>
														<legend className="font-medium">
															{t("api.profile.keys.confirmTitle", {
																name: k.name,
															})}
														</legend>
														<p>{t("api.profile.keys.confirmBody")}</p>
														<button
															type="button"
															className="mr-3 font-medium text-red-700 underline dark:text-red-400"
															onClick={() =>
																router.delete(
																	`/profile/api/keys/${k.id}`,
																	{
																		onFinish: () =>
																			setConfirming(null),
																	},
																)
															}
														>
															{t("api.profile.keys.revoke")}
														</button>
														<button
															type="button"
															className="underline"
															onClick={() => setConfirming(null)}
														>
															{t("api.profile.keys.cancel")}
														</button>
													</fieldset>
												) : (
													<button
														type="button"
														className="underline"
														onClick={() => setConfirming(k.id)}
													>
														{t("api.profile.keys.revoke")}
													</button>
												))}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</section>
				)}
			</section>
		</>
	);
}
