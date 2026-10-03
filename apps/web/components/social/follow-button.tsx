"use client";

import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { toggleFollow } from "@/actions/follows";
import { buttonClass, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";

export function FollowButton({ nickname, initial }: { nickname: string; initial: boolean }) {
	const t = useTranslations("social.profile");
	const tAction = useTranslations("actions");
	const [following, setFollowing] = useState(initial);
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();

	function onClick(): void {
		const next = !following;
		setFollowing(next);
		setError(null);
		startTransition(async () => {
			const result = await toggleFollow({ nickname, follow: next });
			if (!result.ok) {
				setFollowing(!next);
				setError(result.error);
			}
		});
	}

	return (
		<div className="grid gap-2">
			<button
				type="button"
				aria-pressed={following}
				disabled={pending}
				onClick={onClick}
				className={buttonClass(following ? "secondary" : "primary")}
			>
				{following ? t("unfollow") : t("follow")}
			</button>
			{error ? <Notice tone="error">{tAction(error)}</Notice> : null}
		</div>
	);
}
