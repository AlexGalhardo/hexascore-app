"use client";

import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { setNickname } from "@/actions/account";
import { buttonClass, Field, Input, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import { NICKNAME_MAX, normalizeNickname } from "@/lib/validation";

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
			<Field id="nickname" label={ta("fields.nickname")} hint={ta("signup.nicknameHint")}>
				<Input
					id="nickname"
					required
					autoComplete="username"
					autoCapitalize="none"
					spellCheck={false}
					maxLength={NICKNAME_MAX}
					value={nickname}
					aria-invalid={error === "invalid" || error === "nicknameTaken"}
					aria-describedby="nickname-hint"
					onChange={(e) => setValue(normalizeNickname(e.target.value))}
				/>
			</Field>
			<button type="submit" disabled={pending || !nickname.trim()} className={buttonClass("primary")}>
				{t("submit")}
			</button>
		</form>
	);
}
