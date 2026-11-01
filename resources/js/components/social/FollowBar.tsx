import { router } from "@inertiajs/react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import Btn from "../ui/Btn";

/** The `social` prop of User/Profile (SDD 14 §12.2); absent while the feature flag is off. */
export interface SocialProps {
	followers?: number;
	following?: number;
	state?: "" | "accepted" | "pending";
}

interface Props {
	nickname: string;
	social: SocialProps;
	canFollow: boolean;
}

export default function FollowBar({ nickname, social, canFollow }: Props) {
	const t = useT();
	const l = useLocalized();
	const url = l(`/u/${nickname.toLowerCase()}/follow`);
	const options = { preserveScroll: true };

	return (
		<section
			aria-label={t("social.follow.label")}
			className="mt-6 flex flex-wrap items-center gap-x-6 gap-y-3 border-y border-border py-3"
		>
			{social.followers !== undefined && (
				<span className="text-[15px] font-semibold tabular-nums">
					{t("social.follow.followers", { count: social.followers })}
				</span>
			)}
			{social.following !== undefined && (
				<span className="text-[15px] font-semibold tabular-nums">
					{t("social.follow.following", { count: social.following })}
				</span>
			)}
			{canFollow && (
				<span className="ml-auto">
					{social.state ? (
						<Btn kind="secondary" onClick={() => router.delete(url, options)}>
							{t(
								social.state === "pending"
									? "social.follow.requested"
									: "social.follow.unfollow",
							)}
						</Btn>
					) : (
						<Btn onClick={() => router.put(url, {}, options)}>
							{t("social.follow.follow")}
						</Btn>
					)}
				</span>
			)}
		</section>
	);
}
