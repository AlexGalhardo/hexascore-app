import { Link, router, usePage } from "@inertiajs/react";
import { type ReactNode, useEffect, useState } from "react";
import CookieBanner from "../components/CookieBanner";
import Flash from "../components/Flash";
import { localized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

// The sign-in pages stand alone, without the site navbar and footer (owner 2026-11-07).
export const authPages = new Set([
	"Auth/Login",
	"Auth/Signup",
	"Auth/TwoFactorChallenge",
	"Auth/Recover",
	"Auth/Reset",
	"Auth/CheckEmail",
]);

export default function AuthLayout({ children }: { children: ReactNode }) {
	const t = useT();
	const page = usePage<SharedProps>();
	const [announcement, setAnnouncement] = useState("");

	useEffect(() => {
		document.documentElement.lang = page.props.locale;
	}, [page.props.locale]);

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

	const flash = page.flash as { status?: string; success?: string; error?: string } | undefined;
	const success = flash?.success ?? flash?.status;

	return (
		<>
			{/* A way back to the site, top left, as on Resend's sign-in pages. */}
			<Link
				href={localized("/", page.props.locale)}
				className="absolute top-4 left-4 inline-flex min-h-11 items-center gap-1 rounded-md px-3 text-sm text-muted hover:bg-hover hover:text-fg"
			>
				<span aria-hidden="true">‹</span>
				{t("nav.home")}
			</Link>
			<main id="main" className="shell flex min-h-dvh flex-col justify-center py-12">
				{children}
			</main>
			{typeof success === "string" && (
				<Flash message={success} kind="status" position="bottom" />
			)}
			{typeof flash?.error === "string" && (
				<Flash message={flash.error} kind="error" position="bottom" />
			)}
			<div aria-live="polite" role="status" className="sr-only">
				{announcement}
			</div>
			<CookieBanner />
		</>
	);
}
