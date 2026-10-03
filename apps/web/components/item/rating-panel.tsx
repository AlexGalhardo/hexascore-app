import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/navigation";

type Props = { itemId: string };

/**
 * SLOT — milestone 0.3 replaces this with the 6-category rating form (slider + number input per category,
 * live average, verified e-mail required). Until then it only links to login.
 */
export async function RatingPanel({ itemId }: Props) {
	const t = await getTranslations("item");
	return (
		<section aria-label={t("rate")} data-slot="rating-panel" data-item-id={itemId}>
			<Link
				href="/login"
				className="inline-flex min-h-11 items-center rounded-md bg-fg px-4 text-sm font-medium text-bg"
			>
				{t("loginToRate")}
			</Link>
		</section>
	);
}
