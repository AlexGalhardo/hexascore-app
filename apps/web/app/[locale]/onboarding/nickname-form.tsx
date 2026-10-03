"use client";

import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { setNickname } from "@/actions/account";
import { buttonClass, Field, Input, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import { NICKNAME_MAX } from "@/lib/validation";

export function NicknameForm({ next }: { next: string }) {
	const t = useTranslations("social.onboarding");
	const ta = useTranslations("auth");
	const tAction = useTranslations("actions");
	const [nickname, setValue] = useState("");
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();

	function submit(): void {
		setError(null);
		startTransition(async () => {
			const result = await setNickname(nickname);
			if (!result.ok) return setError(result.error);
			// Full navigation so the navbar and session-derived UI pick up the new nickname.
			window.location.assign(next);
		});
	}

	return (
		<form
			className="flex flex-col gap-4"
			onSubmit={(e) => {
				e.preventDefault();
				submit();
			}}
		>
			{error ? <Notice tone="error">{tAction(error)}</Notice> : null}
			<div className="flex items-end gap-3">
				<div
					role="img"
					aria-label={ta("signup.avatar")}
					className="grid size-14 shrink-0 place-items-center rounded-full bg-surface font-serif text-2xl font-bold ring-1 ring-border"
				>
					{nickname.trim().charAt(0).toUpperCase() || "?"}
				</div>
				<div className="flex-1">
					<Field id="nickname" label={ta("fields.nickname")} hint={ta("signup.nicknameHint")}>
						<Input
							id="nickname"
							required
							autoComplete="username"
							maxLength={NICKNAME_MAX}
							value={nickname}
							aria-invalid={error === "invalid" || error === "nicknameTaken"}
							aria-describedby="nickname-hint"
							onChange={(e) => setValue(e.target.value)}
						/>
					</Field>
				</div>
			</div>
			<button type="submit" disabled={pending || !nickname.trim()} className={buttonClass("primary")}>
				{t("submit")}
			</button>
		</form>
	);
}
