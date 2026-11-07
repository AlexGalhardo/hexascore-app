import Constants from "expo-constants";
import type { ReactElement } from "react";
import { About } from "../src/profile/About";
import { useLocale } from "../src/providers/root";
import { Screen } from "../src/ui/Screen";

export default function AboutScreen(): ReactElement {
	return (
		<Screen>
			<About locale={useLocale()} version={Constants.expoConfig?.version ?? "0.0.0"} />
		</Screen>
	);
}
