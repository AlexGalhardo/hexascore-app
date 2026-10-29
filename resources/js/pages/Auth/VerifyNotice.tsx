import { Head, Link, usePage } from "@inertiajs/react";

import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

export default function VerifyNotice() {
	const t = useT();
	const { auth } = usePage<SharedProps>().props;

	return (
		<>
			<Head title={t("auth.verify.metaTitle")} />
			<h1 className="mb-2 text-2xl font-semibold">{t("auth.verify.title")}</h1>
			<p className="mb-4">{t("auth.verify.sentTo", { email: auth?.user?.email ?? "" })}</p>
			<Link
				href="/verify-email/resend"
				method="post"
				as="button"
				className="rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500"
			>
				{t("auth.verify.resend")}
			</Link>
		</>
	);
}
