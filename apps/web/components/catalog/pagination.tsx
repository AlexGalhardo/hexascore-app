import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/navigation";

type Props = {
	pathname: string;
	query: Record<string, string>;
	prevCursor: string | null;
	nextCursor: string | null;
};

const linkClass = "inline-flex min-h-11 items-center rounded-md border border-border px-4 text-sm hover:bg-surface";

export async function Pagination({ pathname, query, prevCursor, nextCursor }: Props) {
	if (!prevCursor && !nextCursor) return null;
	const t = await getTranslations("catalog");
	return (
		<nav aria-label={t("pagination")} className="mt-4 flex justify-between gap-2">
			{prevCursor ? (
				<Link href={{ pathname, query: { ...query, cursor: prevCursor } }} rel="prev" className={linkClass}>
					← {t("previous")}
				</Link>
			) : (
				<span />
			)}
			{nextCursor ? (
				<Link href={{ pathname, query: { ...query, cursor: nextCursor } }} rel="next" className={linkClass}>
					{t("next")} →
				</Link>
			) : null}
		</nav>
	);
}
