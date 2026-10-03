"use client";

import { useTranslations } from "next-intl";
import { useEffect } from "react";

type Props = { error: Error & { digest?: string }; reset: () => void };

export default function LocaleError({ error, reset }: Props) {
	const t = useTranslations("errors");

	useEffect(() => {
		console.error("[ui] unhandled error", error);
	}, [error]);

	return (
		<section className="max-w-xl py-12" role="alert">
			<p className="font-mono text-sm text-muted">500</p>
			<h1 className="mt-2 font-serif text-3xl font-bold">{t("serverTitle")}</h1>
			<p className="mt-3 text-muted">{t("serverBody")}</p>
			{error.digest ? (
				<p className="mt-2 font-mono text-xs text-muted">
					{t("reference")}: {error.digest}
				</p>
			) : null}
			<button
				type="button"
				onClick={reset}
				className="mt-6 inline-flex min-h-11 items-center rounded-md bg-fg px-4 text-bg"
			>
				{t("retry")}
			</button>
		</section>
	);
}
