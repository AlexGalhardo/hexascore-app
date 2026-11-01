import type { ReactElement } from "react";
import { ScrollView, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";

const FG = "font-text text-16 text-fg dark:text-fg-dark";
const SECTION = "gap-3 border-t-2 border-fg pt-4 dark:border-fg-dark";

/** FR-10-9: the About screen carries the TMDB and IGDB attribution their terms require. */
export function About({ locale, version }: { locale: AppLocale; version: string }): ReactElement {
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-4 px-4 pb-8 pt-6"
		>
			<Text
				accessibilityRole="header"
				className="font-display text-36 font-bold text-fg dark:text-fg-dark"
			>
				{t(locale, "account.about.title")}
			</Text>
			<View className={SECTION}>
				<Text className="font-display text-24 font-bold text-fg dark:text-fg-dark">HexaScore</Text>
				<Text className="font-text text-14 tabular-nums text-muted dark:text-muted-dark">
					{t(locale, "account.about.version", { version })}
				</Text>
			</View>
			<View testID="about-sources" className={SECTION}>
				<Text testID="attribution-tmdb" className={FG}>
					{t(locale, "account.about.tmdb")}
				</Text>
				<Text testID="attribution-igdb" className={FG}>
					{t(locale, "account.about.igdb")}
				</Text>
			</View>
		</ScrollView>
	);
}
