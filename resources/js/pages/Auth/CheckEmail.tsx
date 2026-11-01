import { Head, Link } from "@inertiajs/react";

import { useT } from "../../lib/useT";

export default function CheckEmail() {
	const t = useT();

	return (
		<>
			<Head title={t("auth.signup.checkTitle")} />
			<h1 className="mb-2 text-2xl font-semibold">{t("auth.signup.checkTitle")}</h1>
			<p className="mb-4">{t("auth.signup.checkNotice")}</p>
			<Link href="/login" className="text-primary hover:underline">
				{t("auth.signup.goLogin")}
			</Link>
		</>
	);
}
