import { useQuery, useQueryClient } from "@tanstack/react-query";
import { type ReactElement, useState } from "react";
import { Text, View } from "react-native";
import { describeError } from "../api/problems";
import { type AppLocale, t } from "../i18n";
import { Btn } from "../ui/Btn";
import { SettingsSection } from "../ui/screens/SettingsPage";
import { type Device, devicesKey, type ProfileApi, parseDevices } from "./me";

/** Profile → Devices: GET /tokens, DELETE /tokens/{id} then refetch (SDD 10 §6.2). */
export function Devices({
	api,
	locale,
	onCurrentRevoked,
}: {
	api: ProfileApi;
	locale: AppLocale;
	/** Revoking this device's own token ends the session. */
	onCurrentRevoked: () => void;
}): ReactElement {
	const queryClient = useQueryClient();
	const devices = useQuery({
		queryKey: devicesKey,
		queryFn: async () => parseDevices(await api.read.tokensIndex()),
	});
	const [error, setError] = useState<string | null>(null);

	const revoke = async (device: Device): Promise<void> => {
		setError(null);
		try {
			await api.write((client) => client.tokensRevoke({ params: { id: device.id } }));
		} catch (cause) {
			// The current device signs out anyway: the local wipe protects the user (as logout does).
			if (!device.current) {
				setError(describeError(cause).key);
				return;
			}
		}
		if (device.current) onCurrentRevoked();
		else await queryClient.invalidateQueries({ queryKey: devicesKey });
	};

	return (
		<SettingsSection title={t(locale, "account.sessions.title")}>
			{(devices.data ?? []).map((device) => (
				<View key={device.id} testID={`device-${device.id}`} className="gap-2">
					<Text className="font-text text-16 font-semibold text-fg dark:text-fg-dark">
						{device.name}
						{device.current ? ` · ${t(locale, "account.sessions.current")}` : ""}
					</Text>
					{device.lastUsedAt ? (
						<Text className="font-text text-14 text-muted dark:text-muted-dark">
							{`${t(locale, "account.sessions.lastSeen")}: ${new Date(device.lastUsedAt).toLocaleDateString(locale)}`}
						</Text>
					) : null}
					<Btn
						variant="secondary"
						label={t(locale, "account.sessions.revoke")}
						accessibilityHint={device.name}
						onPress={() => void revoke(device)}
					/>
				</View>
			))}
			{error ? (
				<Text accessibilityRole="alert" className="font-text text-14 text-fg dark:text-fg-dark">
					{t(locale, error)}
				</Text>
			) : null}
		</SettingsSection>
	);
}
