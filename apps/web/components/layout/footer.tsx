import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/navigation";
import { APP_NAME } from "@/lib/config";

export async function Footer() {
	const t = await getTranslations("footer");

	return (
		<footer className="mt-16 border-t border-border">
			<div className="mx-auto flex max-w-6xl flex-wrap items-center gap-4 px-4 py-6 text-sm text-muted">
				<span className="mr-auto">© 2026 {APP_NAME}</span>
				<Link href="/terms">{t("terms")}</Link>
				<Link href="/policy">{t("privacy")}</Link>
				<Link href="/contact">{t("contact")}</Link>
			</div>
		</footer>
	);
}
