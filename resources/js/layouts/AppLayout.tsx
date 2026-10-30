import { Link, router, usePage } from "@inertiajs/react";
import { type ReactNode, useEffect, useState } from "react";
import Flash from "../components/Flash";
import Footer from "../components/Footer";
import Nav from "../components/Nav";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

const link = "text-accent-game hover:underline";

export default function AppLayout({ children }: { children: ReactNode }) {
	const t = useT();
	const page = usePage<SharedProps>();
	const { auth } = page.props;
	const [announcement, setAnnouncement] = useState("");

	// Client-side visits keep the server-rendered <html>, so follow the locale of the page (SDD 09 §7.6).
	useEffect(() => {
		document.documentElement.lang = page.props.locale;
	}, [page.props.locale]);

	// After a visit (not a partial reload) focus moves to the new <h1> and the
	// title is announced politely (SDD 09 §7.6).
	useEffect(
		() =>
			router.on("navigate", () => {
				const h1 = document.querySelector<HTMLElement>("h1");
				h1?.setAttribute("tabindex", "-1");
				h1?.focus();
				setAnnouncement(document.title);
			}),
		[],
	);

	const scheduled = auth?.user?.deletion_scheduled_at;
	const deletionDate = scheduled
		? new Date(new Date(scheduled).getTime() + 30 * 86_400_000).toLocaleDateString(
				page.props.locale,
			)
		: null;
	const flash = page.flash as { status?: string; success?: string; error?: string } | undefined;
	const success = flash?.success ?? flash?.status;

	return (
		<>
			<a
				href="#main"
				className="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-[80] focus:rounded-md focus:bg-bg focus:px-4 focus:py-2"
			>
				{t("nav.skip")}
			</a>
			<Nav />
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
			<main id="main" className="mx-auto max-w-6xl p-4">
				{children}
			</main>
			<Footer />
			{typeof success === "string" && <Flash message={success} kind="status" />}
			{typeof flash?.error === "string" && <Flash message={flash.error} kind="error" />}
			<div aria-live="polite" role="status" className="sr-only">
				{announcement}
			</div>
		</>
	);
}
