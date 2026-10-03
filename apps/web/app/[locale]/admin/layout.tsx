import type { Metadata } from "next";
import type { ReactNode } from "react";
import { Link } from "@/i18n/navigation";
import { requireAdmin } from "./guard";

export const metadata: Metadata = { title: "Admin", robots: { index: false, follow: false } };

const LINKS = [
	["/admin", "Overview"],
	["/admin/reports", "Reports"],
	["/admin/contact", "Contact"],
	["/admin/items", "Items"],
	["/admin/ingestion", "Ingestion"],
] as const;

// UI is English-only (no messages), but served under any locale prefix: next-intl redirects /admin to the
// visitor's cookie locale, so a locale 404 would lock out admins browsing in pt-BR/es.
// Non-admins get the regular 404 so the area is not revealed. Pages re-check (a layout is not a security
// boundary on client navigation) and every admin action re-checks on its own.
export default async function AdminLayout({ children }: { children: ReactNode }) {
	await requireAdmin();
	return (
		<div lang="en" className="grid gap-6">
			<nav aria-label="Admin" className="flex flex-wrap gap-1 border-b border-border text-sm">
				{LINKS.map(([href, label]) => (
					<Link key={href} href={href} className="grid min-h-11 place-items-center px-3 hover:bg-surface">
						{label}
					</Link>
				))}
			</nav>
			{children}
		</div>
	);
}
