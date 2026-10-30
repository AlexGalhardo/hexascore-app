import { useT } from "../lib/useT";

// ChallengeStatus announces the proof-of-work check (SDD 13 §7.2) and, without JavaScript, points to
// "Continue with Google", which is never challenged.
export default function ChallengeStatus({ checking }: { checking: boolean }) {
	const t = useT();
	return (
		<>
			<p aria-live="polite" className="text-sm text-zinc-600 dark:text-zinc-400">
				{checking ? t("auth.challenge.checking") : ""}
			</p>
			<noscript>
				<p className="text-sm">
					{t("auth.challenge.noscript")}{" "}
					<a href="/auth/google/redirect" className="underline">
						{t("auth.google")}
					</a>
				</p>
			</noscript>
		</>
	);
}
