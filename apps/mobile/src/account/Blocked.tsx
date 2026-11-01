import type { ReactElement } from "react";
import { type AppLocale, t } from "../i18n";
import { StatePage } from "../ui/screens/StatePage";

// UTC so the server's end date is not shifted to the previous day west of Greenwich.
const longDate = (locale: AppLocale, iso: string): string =>
	new Intl.DateTimeFormat(locale, { dateStyle: "long", timeZone: "UTC" }).format(new Date(iso));

/** 403 `account-banned` (SDD 10 §6.2): the vault is already wiped; leaving returns to login. */
export function BannedBody(props: {
	locale: AppLocale;
	reason: string | null;
	expiresAt: string | null;
	onLeave: () => void;
}): ReactElement {
	const { locale } = props;
	return (
		<StatePage
			title={t(locale, "account.blocked.title")}
			body={[
				props.reason ? t(locale, "account.blocked.reason", { reason: props.reason }) : null,
				props.expiresAt
					? t(locale, "account.blocked.until", { date: longDate(locale, props.expiresAt) })
					: t(locale, "account.blocked.permanent"),
			].filter((line): line is string => line !== null)}
			actions={[{ label: t(locale, "account.blocked.back"), onPress: props.onLeave }]}
		/>
	);
}

/** `deletionScheduledAt` on `/me`: cancelling happens on web `/login` (SDD 10 §6.2, SDD 03 §8). */
export function DeletionScheduledBody(props: {
	locale: AppLocale;
	scheduledAt: string;
	openWeb: (path: string) => void;
	onLogout: () => void;
}): ReactElement {
	const { locale } = props;
	return (
		<StatePage
			title={t(locale, "account.banner.text", { date: longDate(locale, props.scheduledAt) })}
			body={[]}
			actions={[
				{ label: t(locale, "account.banner.cancel"), onPress: () => props.openWeb("/login") },
				{
					label: t(locale, "account.blocked.back"),
					onPress: props.onLogout,
					variant: "secondary",
				},
			]}
		/>
	);
}
