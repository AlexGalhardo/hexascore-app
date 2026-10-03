"use client";

import { useTranslations } from "next-intl";
import { type ComponentProps, useState } from "react";
import { inputClass } from "@/components/ui/form";

export function PasswordInput(props: Omit<ComponentProps<"input">, "type">) {
	const t = useTranslations("auth.fields");
	const [visible, setVisible] = useState(false);

	return (
		<div className="relative">
			<input {...props} type={visible ? "text" : "password"} className={`${inputClass} pr-12`} />
			<button
				type="button"
				onClick={() => setVisible((v) => !v)}
				aria-label={visible ? t("hide") : t("show")}
				aria-pressed={visible}
				aria-controls={props.id}
				className="absolute inset-y-0 right-0 grid w-11 place-items-center text-muted hover:text-fg"
			>
				<svg
					aria-hidden="true"
					viewBox="0 0 24 24"
					className="size-5"
					fill="none"
					stroke="currentColor"
					strokeWidth="2"
				>
					<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z" />
					<circle cx="12" cy="12" r="3" />
					{visible && <path d="M3 3l18 18" />}
				</svg>
			</button>
		</div>
	);
}
