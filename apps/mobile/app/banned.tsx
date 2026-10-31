import type { ReactElement } from "react";
import { BannedBody } from "../src/account/Blocked";
import { session } from "../src/auth/runtime";
import { useLocale } from "../src/providers/root";

export default function BannedScreen(): ReactElement {
	const locale = useLocale();
	const state = session.state();
	const ban = state.kind === "banned" ? state : { reason: null, expiresAt: null };
	return (
		<BannedBody
			locale={locale}
			reason={ban.reason}
			expiresAt={ban.expiresAt}
			onLeave={() => void session.signOut()}
		/>
	);
}
