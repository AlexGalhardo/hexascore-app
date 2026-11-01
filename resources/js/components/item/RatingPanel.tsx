import { Link, router, usePage } from "@inertiajs/react";
import { type ReactNode, useState } from "react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import StarRating from "../rating/StarRating";
import Btn, { btnClass } from "../ui/Btn";
import ScoreBadge from "../ui/ScoreBadge";
import { CATEGORY_KEYS, type ItemType } from "./types";

interface Props {
	itemId: string;
	title: string;
	type: ItemType;
	mine: Record<string, number> | null | undefined;
	canWrite: boolean;
	children?: ReactNode;
}

// The viewer's six-category rating (design v1 "Your rating" panel). Values are local until saved;
// the server needs every category, so Save stays disabled until all six are set.
export default function RatingPanel({ itemId, title, type, mine, canWrite, children }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const l = (path: string) => localized(path, locale);
	const keys = CATEGORY_KEYS[type];
	const [values, setValues] = useState<Record<string, number | null>>(
		Object.fromEntries(keys.map((k) => [k, mine?.[k] ?? null])),
	);
	const [busy, setBusy] = useState(false);
	const rated = keys.map((k) => values[k]).filter((v): v is number => v !== null);
	const complete = rated.length === keys.length;
	// Preview only; the server computes the stored overall itself.
	const overall = rated.length
		? Math.round((rated.reduce((a, b) => a + b, 0) / rated.length) * 10) / 10
		: null;

	const visit = { preserveScroll: true, onFinish: () => setBusy(false) };
	const save = () => {
		setBusy(true);
		router.put(l(`/items/${itemId}/rating`), { scores: values }, visit);
	};
	const remove = () => {
		setBusy(true);
		setValues(Object.fromEntries(keys.map((k) => [k, null])));
		router.delete(l(`/items/${itemId}/rating`), visit);
	};

	return (
		<section
			id="your-rating"
			aria-labelledby="rating-heading"
			className="grid scroll-mt-20 gap-4 border-t-2 border-primary bg-surface p-5"
		>
			<div className="grid gap-2">
				<p className="text-xs font-semibold tracking-widest uppercase">{t("item.rate")}</p>
				<h2 id="rating-heading" className="text-2xl leading-tight">
					{t("item.panel.title", { title })}
				</h2>
				<p className="text-sm text-muted">
					{canWrite ? t("item.panel.help") : t("item.panel.helpGuest")}
				</p>
			</div>
			<div className="grid gap-2">
				{keys.map((k) => (
					<StarRating
						key={k}
						label={t(`categories.${type}.${k}`)}
						value={values[k]}
						onChange={(v) => setValues((s) => ({ ...s, [k]: v }))}
						disabled={!canWrite || busy}
					/>
				))}
			</div>
			{canWrite ? (
				<>
					<div className="flex items-center justify-between gap-4">
						<div className="grid gap-0.5">
							<p className="text-sm font-semibold">{t("item.panel.overall")}</p>
							<p className="text-xs text-muted">
								{t("item.panel.overallHint", {
									rated: rated.length,
									total: keys.length,
								})}
								{!complete && ` ${t("item.panel.completeHint")}`}
							</p>
						</div>
						<ScoreBadge score={overall} size="md" />
					</div>
					<div className="grid gap-2">
						<Btn disabled={!complete || busy} onClick={save} className="w-full">
							{busy
								? t("ratings.saving")
								: mine
									? t("ratings.update")
									: t("ratings.save")}
						</Btn>
						{mine && (
							<Btn kind="ghost" disabled={busy} onClick={remove} className="w-full">
								{t("ratings.delete")}
							</Btn>
						)}
					</div>
					{children}
				</>
			) : (
				<div className="grid gap-2">
					<Link href={l("/login")} className={btnClass("primary", "w-full")}>
						{t("item.loginToRate")}
					</Link>
					<p className="text-sm text-muted">
						{t("item.panel.noAccount")}{" "}
						<Link
							href={l("/signup")}
							className="font-semibold text-primary hover:underline"
						>
							{t("item.panel.signupFree")}
						</Link>
					</p>
				</div>
			)}
		</section>
	);
}
