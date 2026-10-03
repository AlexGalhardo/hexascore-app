"use client";

import { useRouter } from "next/navigation";
import { useFormatter, useTranslations } from "next-intl";
import { useState } from "react";
import { buttonClass, Notice } from "@/components/ui/form";
import { authClient } from "@/lib/auth-client";
import type { AuthClientError } from "@/lib/auth-errors";

type Passkey = { id: string; name: string | null; createdAt: string | null };

export function PasskeysSection({ passkeys }: { passkeys: Passkey[] }) {
	const t = useTranslations("account.passkeys");
	const tAuth = useTranslations("auth");
	const format = useFormatter();
	const router = useRouter();
	const [error, setError] = useState<string | null>(null);
	const [pending, setPending] = useState(false);

	async function run(fn: () => Promise<{ error: AuthClientError } | undefined>, failure: string): Promise<void> {
		setPending(true);
		setError(null);
		try {
			const result = await fn();
			if (result?.error) setError(result.error.status === 429 ? tAuth("errors.rateLimited") : failure);
			else router.refresh();
		} catch {
			setError(failure);
		} finally {
			setPending(false);
		}
	}

	return (
		<div className="flex flex-col gap-4">
			<p className="text-sm text-muted">{t("body")}</p>
			{error && <Notice tone="error">{error}</Notice>}
			{passkeys.length === 0 ? (
				<p className="text-sm">{t("empty")}</p>
			) : (
				<ul className="divide-y divide-border rounded-md border border-border">
					{passkeys.map((p) => {
						const name = p.name || t("unnamed");
						return (
							<li key={p.id} className="flex items-center justify-between gap-3 px-3 py-2">
								<div className="min-w-0">
									<p className="truncate font-medium">{name}</p>
									{p.createdAt && (
										<p className="text-sm text-muted">
											{t("added", {
												date: format.dateTime(new Date(p.createdAt), { dateStyle: "medium" }),
											})}
										</p>
									)}
								</div>
								<button
									type="button"
									disabled={pending}
									onClick={() =>
										run(
											() => authClient.passkey.deletePasskey({ id: p.id }),
											tAuth("errors.generic"),
										)
									}
									aria-label={t("removeLabel", { name })}
									className={buttonClass("secondary")}
								>
									{t("remove")}
								</button>
							</li>
						);
					})}
				</ul>
			)}
			<button
				type="button"
				disabled={pending}
				onClick={() => run(() => authClient.passkey.addPasskey(), t("failed"))}
				className={buttonClass("secondary", "self-start")}
			>
				{t("add")}
			</button>
		</div>
	);
}
