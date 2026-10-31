import { Link, usePage } from "@inertiajs/react";

import Seo from "../components/Seo";
import { localized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

const card =
	"grid gap-1 rounded-lg border border-border bg-surface p-4 hover:border-fg focus-visible:outline-2 focus-visible:outline-offset-2";

// Landing page after login: the member's own shelves, lists and page.
export default function Dashboard() {
	const t = useT();
	const { auth, locale } = usePage<SharedProps>().props;
	const user = auth?.user;
	const l = (path: string) => localized(path, locale);

	const cards = [
		{ href: l("/library"), title: t("dashboard.library"), help: t("dashboard.libraryHelp") },
		{ href: l("/lists"), title: t("dashboard.lists"), help: t("dashboard.listsHelp") },
		...(user?.nickname
			? [
					{
						href: l(`/u/${user.nickname.toLowerCase()}`),
						title: t("dashboard.page"),
						help: t("dashboard.pageHelp"),
					},
				]
			: []),
		{ href: "/profile/api", title: t("dashboard.api"), help: t("dashboard.apiHelp") },
	];

	return (
		<>
			<Seo title={t("dashboard.title", { name: user?.name ?? "" })} noindex />
			<h1 className="text-2xl font-semibold">
				{t("dashboard.title", { name: user?.name ?? "" })}
			</h1>
			<p className="mt-1 text-muted">{t("dashboard.lead")}</p>
			{user && !user.email_verified_at && (
				<p className="mt-4">
					<Link href="/verify-email" className="underline">
						{t("dashboard.verify")}
					</Link>
				</p>
			)}
			<div className="mt-6 grid gap-3 sm:grid-cols-2">
				{cards.map((c) => (
					<Link key={c.href} href={c.href} className={card}>
						<span className="font-semibold">{c.title}</span>
						<span className="text-sm text-muted">{c.help}</span>
					</Link>
				))}
			</div>
			<p className="mt-6 text-sm">
				<Link href="/user/two-factor" className="underline">
					{t("dashboard.twoFactor")}
				</Link>
			</p>
		</>
	);
}
