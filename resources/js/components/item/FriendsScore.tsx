import { useT } from "../../lib/useT";
import ScoreBadge from "../ui/ScoreBadge";

export interface FriendsScoreData {
	count: number;
	average: number;
	friends: { nickname: string; score: number }[];
}

// Next to the community score: the average of the people the viewer follows (SDD 14 §12).
export default function FriendsScore({ data }: { data: FriendsScoreData | null | undefined }) {
	const t = useT();
	if (!data) return null;

	return (
		<div
			data-testid="friends-score"
			className="flex items-center gap-4 border border-border bg-surface p-4"
		>
			<ScoreBadge score={data.average} size="md" />
			<div className="grid gap-1">
				<p className="font-semibold">{t("friendsScore.title")}</p>
				<p className="text-sm text-muted">
					{t("friendsScore.summary", { count: data.count })}
					{": "}
					{data.friends.map((f) => `${f.nickname} ${f.score.toFixed(1)}`).join(", ")}
				</p>
			</div>
		</div>
	);
}
