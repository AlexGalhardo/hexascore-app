import { type ReactElement, useEffect, useState } from "react";
import { Text, View } from "react-native";
import type { Client } from "../../../../packages/app-client/src/api";
import { describeError } from "../api/problems";
import { type AppLocale, t } from "../i18n";
import { Btn } from "../ui/Btn";
import { SettingsSection } from "../ui/screens/SettingsPage";

export type DeletionWrite = <T>(
	call: (client: Pick<Client, "meDeletionSchedule">) => Promise<T>,
) => Promise<T>;

export const COUNTDOWN_S = 10;

const FG = "font-text text-14 text-fg dark:text-fg-dark";

/** FR-10-8: POST /me/deletion only after a 10 s countdown (FR-03-15). */
export function DeleteAccount({
	locale,
	write,
	onScheduled,
}: {
	locale: AppLocale;
	write: DeletionWrite;
	/** Gets `deletesAt`; the caller routes to the deletion-scheduled screen. */
	onScheduled: (deletesAt: string) => void;
}): ReactElement {
	const [until, setUntil] = useState<number | null>(null);
	const [now, setNow] = useState(() => Date.now());
	const [error, setError] = useState<string | null>(null);
	const left = until === null ? null : Math.max(0, Math.ceil((until - now) / 1000));

	// Counted from a deadline, so a late tick (busy JS thread) never stretches the 10 s.
	useEffect(() => {
		if (until === null) return;
		const id = setInterval(() => setNow(Date.now()), 250);
		return () => clearInterval(id);
	}, [until]);

	const open = (): void => {
		const start = Date.now();
		setNow(start);
		setUntil(start + COUNTDOWN_S * 1000);
	};

	const confirm = async (): Promise<void> => {
		setError(null);
		try {
			const body = await write((client) => client.meDeletionSchedule({ body: { confirm: true } }));
			const at = (body as { deletesAt?: unknown } | null)?.deletesAt;
			onScheduled(typeof at === "string" ? at : new Date().toISOString());
		} catch (cause) {
			setError(describeError(cause).key);
		}
	};

	return (
		<SettingsSection title={t(locale, "account.delete.title")}>
			<Text className={FG}>{t(locale, "account.delete.body")}</Text>
			{left === null ? (
				<Btn variant="secondary" label={t(locale, "account.delete.button")} onPress={open} />
			) : (
				<View
					testID="delete-dialog"
					accessibilityRole="alert"
					className="gap-3 border-t-2 border-danger bg-surface p-4 dark:border-danger-dark dark:bg-surface-dark"
				>
					<Text
						accessibilityRole="header"
						className="font-display text-20 font-bold text-fg dark:text-fg-dark"
					>
						{t(locale, "account.delete.dialogTitle")}
					</Text>
					<Text className={FG}>{t(locale, "account.delete.dialogBody")}</Text>
					<Btn
						label={
							left > 0
								? t(locale, "account.delete.wait", { seconds: left })
								: t(locale, "account.delete.confirm")
						}
						disabled={left > 0}
						onPress={() => void confirm()}
					/>
					<Btn
						variant="ghost"
						label={t(locale, "account.delete.cancel")}
						onPress={() => setUntil(null)}
					/>
				</View>
			)}
			{error ? (
				<Text accessibilityRole="alert" className={FG}>
					{t(locale, error)}
				</Text>
			) : null}
		</SettingsSection>
	);
}
