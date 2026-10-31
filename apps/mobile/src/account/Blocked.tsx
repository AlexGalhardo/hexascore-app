import type { ReactElement, ReactNode } from "react";
import { Pressable, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";

// UTC so the server's end date is not shifted to the previous day west of Greenwich.
const longDate = (locale: AppLocale, iso: string): string =>
	new Intl.DateTimeFormat(locale, { dateStyle: "long", timeZone: "UTC" }).format(new Date(iso));

function Screen({ children }: { children: ReactNode }): ReactElement {
	return (
		<View className="flex-1 items-center justify-center gap-4 bg-bg px-4 dark:bg-bg-dark">
			{children}
		</View>
	);
}

function Action({ label, onPress }: { label: string; onPress: () => void }): ReactElement {
	return (
		<Pressable accessibilityRole="button" className="min-h-11 justify-center" onPress={onPress}>
			<Text className="text-fg dark:text-fg-dark">{label}</Text>
		</Pressable>
	);
}

/** 403 `account-banned` (SDD 10 §6.2): the vault is already wiped; leaving returns to login. */
export function BannedBody(props: {
	locale: AppLocale;
	reason: string | null;
	expiresAt: string | null;
	onLeave: () => void;
}): ReactElement {
	const { locale } = props;
	return (
		<Screen>
			<Text className="text-20 text-fg dark:text-fg-dark">
				{t(locale, "account.blocked.title")}
			</Text>
			{props.reason ? (
				<Text className="text-fg dark:text-fg-dark">
					{t(locale, "account.blocked.reason", { reason: props.reason })}
				</Text>
			) : null}
			<Text className="text-fg dark:text-fg-dark">
				{props.expiresAt
					? t(locale, "account.blocked.until", { date: longDate(locale, props.expiresAt) })
					: t(locale, "account.blocked.permanent")}
			</Text>
			<Action label={t(locale, "account.blocked.back")} onPress={props.onLeave} />
		</Screen>
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
		<Screen>
			<Text className="text-fg dark:text-fg-dark">
				{t(locale, "account.banner.text", { date: longDate(locale, props.scheduledAt) })}
			</Text>
			<Action label={t(locale, "account.banner.cancel")} onPress={() => props.openWeb("/login")} />
			<Action label={t(locale, "account.blocked.back")} onPress={props.onLogout} />
		</Screen>
	);
}
