import { type ReactElement, useState } from "react";
import { Text } from "react-native";
import { ApiError, type Client } from "../../../../packages/app-client/src/api";
import { type AppLocale, t } from "../i18n";
import { Btn } from "../ui/Btn";
import { SettingsSection } from "../ui/screens/SettingsPage";

export type ExportWrite = <T>(call: (client: Pick<Client, "meExport">) => Promise<T>) => Promise<T>;

type Outcome = { key: string; days?: number } | null;

const FG = "font-text text-14 text-fg dark:text-fg-dark";

/** POST /me/export: 202 mails a ZIP link; 429 `export-cooldown` carries Retry-After (SDD 14 §7.4). */
export function ExportData({
	locale,
	write,
}: {
	locale: AppLocale;
	write: ExportWrite;
}): ReactElement {
	const [outcome, setOutcome] = useState<Outcome>(null);
	const [busy, setBusy] = useState(false);

	const request = async (): Promise<void> => {
		setBusy(true);
		try {
			await write((client) => client.meExport());
			setOutcome({ key: "account.export.requested" });
		} catch (error) {
			if (error instanceof ApiError && error.status === 429) {
				const seconds: unknown = (error as { retryAfter?: unknown }).retryAfter;
				const days = typeof seconds === "number" ? Math.ceil(seconds / 86_400) : undefined;
				setOutcome({ key: "account.export.cooldown", days });
			} else setOutcome({ key: "account.export.failed" });
		} finally {
			setBusy(false);
		}
	};

	return (
		<SettingsSection title={t(locale, "account.export.title")}>
			<Text className={FG}>{t(locale, "account.export.body")}</Text>
			<Btn
				variant="secondary"
				label={t(locale, "account.export.button")}
				onPress={() => void request()}
				disabled={busy}
			/>
			{outcome ? (
				<Text accessibilityLiveRegion="polite" className={FG}>
					{t(locale, outcome.key)}
				</Text>
			) : null}
			{outcome?.days === undefined ? null : (
				<Text className={FG}>{t(locale, "account.export.retryIn", { count: outcome.days })}</Text>
			)}
		</SettingsSection>
	);
}
