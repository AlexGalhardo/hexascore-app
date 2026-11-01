import { Head, Link } from "@inertiajs/react";

import { useT } from "../../lib/useT";

const link = "text-primary hover:underline";

export default function DeleteAccount() {
	const t = useT();

	return (
		<>
			<Head title={t("help.deleteAccount.metaTitle")} />
			<h1 className="mb-4 text-2xl font-semibold">{t("help.deleteAccount.title")}</h1>
			<p className="mb-4">{t("help.deleteAccount.intro")}</p>
			<ol className="mb-4 list-decimal space-y-1 pl-6">
				{[0, 1, 2].map((i) => (
					<li key={i}>{t(`help.deleteAccount.steps.${i}`)}</li>
				))}
			</ol>
			<p className="mb-2">{t("help.deleteAccount.grace")}</p>
			<p className="mb-2">{t("help.deleteAccount.erased")}</p>
			<p className="mb-4">{t("help.deleteAccount.kept")}</p>
			<p className="flex gap-4">
				<Link href="/profile#delete" className={link}>
					{t("help.deleteAccount.cta")}
				</Link>
				<Link href="/login" className={link}>
					{t("help.deleteAccount.login")}
				</Link>
			</p>
		</>
	);
}
