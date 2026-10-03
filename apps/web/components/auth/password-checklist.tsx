"use client";

import { useTranslations } from "next-intl";
import { checkPassword, PASSWORD_RULES } from "@/lib/password-rules";

export function PasswordChecklist({ id, value }: { id: string; value: string }) {
	const t = useTranslations("auth.rules");
	const result = checkPassword(value);

	return (
		<div id={id} className="text-sm">
			<p className="text-muted">{t("title")}</p>
			<ul className="mt-1 grid gap-0.5 sm:grid-cols-2">
				{PASSWORD_RULES.map((rule) => (
					<li key={rule} className={result[rule] ? "text-score-high" : "text-muted"}>
						<span aria-hidden="true" className="inline-block w-5 font-mono">
							{result[rule] ? "✓" : "○"}
						</span>
						{t(rule)}
						<span className="sr-only">{result[rule] ? " ✓" : ""}</span>
					</li>
				))}
			</ul>
		</div>
	);
}
