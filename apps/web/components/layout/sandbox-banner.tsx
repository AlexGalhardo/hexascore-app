import { getTranslations } from "next-intl/server";
import type { ReactNode } from "react";
import { env } from "@/lib/env";

// Not dismissable on purpose: every visitor of the sandbox must know the users, ratings and lists are fake.
export async function SandboxBanner(): Promise<ReactNode> {
	if (!env.NEXT_PUBLIC_SANDBOX) return null;
	const t = await getTranslations("sandbox");
	return (
		<div role="status" className="border-b-2 border-score-mid bg-fg text-bg">
			<p className="mx-auto max-w-6xl px-4 py-2 text-center text-sm font-medium">{t("banner")}</p>
		</div>
	);
}
