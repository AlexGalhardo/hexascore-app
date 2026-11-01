import { router } from "@inertiajs/react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

/** The feed opt-in switch of /profile (SDD 14 FR-14-18); it saves on change. */
export default function FeedOptIn({ on }: { on: boolean }) {
	const t = useT();
	const l = useLocalized();

	return (
		<label className="flex items-start gap-2">
			<input
				type="checkbox"
				checked={on}
				onChange={(e) =>
					router.patch(
						l("/profile/feed"),
						{ feedOptIn: e.target.checked },
						{ preserveScroll: true },
					)
				}
			/>
			<span>
				<span className="font-medium">{t("social.optIn.label")}</span>
				<span className="block text-sm text-muted">{t("social.optIn.help")}</span>
			</span>
		</label>
	);
}
