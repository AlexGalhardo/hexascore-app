import { Link, usePage } from "@inertiajs/react";
import type { ReactNode } from "react";
import { demoInstalled, demoLink } from "../demo";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

const link = "text-indigo-600 hover:underline dark:text-indigo-400";

export default function AppLayout({ children }: { children: ReactNode }) {
	const t = useT();
	const page = usePage<SharedProps>();
	const { auth } = page.props;
	const scheduled = auth?.user?.deletion_scheduled_at;
	const deletionDate = scheduled
		? new Date(new Date(scheduled).getTime() + 30 * 86_400_000).toLocaleDateString(
				page.props.locale as string,
			)
		: null;
	const message = page.flash?.success ?? page.flash?.status;

	return (
		<>
			<header className="flex items-center justify-between border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
				<Link href="/" className="font-bold">
					{page.props.appName}
				</Link>
				<nav className="flex items-center gap-4">
					{demoInstalled && (
						<Link href={demoLink.href} className={link}>
							{demoLink.label}
						</Link>
					)}
					{auth?.user ? (
						<>
							<Link href="/dashboard" className={link}>
								{auth.user.name}
							</Link>
							<Link href="/logout" method="post" as="button" className={link}>
								Log out
							</Link>
						</>
					) : auth ? (
						<>
							<Link href="/login" className={link}>
								Log in
							</Link>
							<Link href="/signup" className={link}>
								Register
							</Link>
						</>
					) : null}
				</nav>
			</header>
			{deletionDate && (
				<div
					role="alert"
					className="flex items-center justify-between gap-4 bg-amber-100 px-4 py-2 text-amber-900"
				>
					<span>{t("account.banner.text", { date: deletionDate })}</span>
					<Link href="/profile/deletion" method="delete" as="button" className={link}>
						{t("account.banner.cancel")}
					</Link>
				</div>
			)}
			<main className="mx-auto max-w-2xl p-4">
				{typeof message === "string" && (
					<p
						role="status"
						className="mb-4 rounded-md bg-indigo-50 px-4 py-3 text-indigo-900 dark:bg-indigo-950 dark:text-indigo-100"
					>
						{message}
					</p>
				)}
				{children}
			</main>
		</>
	);
}
