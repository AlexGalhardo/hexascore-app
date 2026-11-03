import { Link, usePage } from "@inertiajs/react";
import { ArrowLeft, Menu, Moon, Sun } from "lucide-react";
import { type ReactElement, type ReactNode, useState } from "react";

import { activeHref, NAV, type NavItem } from "../components/admin/nav";
import { Avatar } from "../components/admin/ui/avatar";
import { Button } from "../components/admin/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "../components/admin/ui/dropdown-menu";
import { Sheet } from "../components/admin/ui/sheet";
import { BRAND } from "../lib/brand";
import { cn } from "../lib/cn";
import { toggleTheme } from "../lib/theme";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

const ITEM =
	"flex min-h-9 items-center gap-2.5 rounded-md px-2 py-1.5 text-sm text-muted outline-none transition-colors hover:bg-hover hover:text-fg focus-visible:ring-2 focus-visible:ring-primary aria-[current=page]:bg-surface aria-[current=page]:font-medium aria-[current=page]:text-fg [&_svg]:size-4 [&_svg]:shrink-0";

function NavLink({ item, current }: { item: NavItem; current: boolean }): ReactElement {
	const t = useT();
	const Icon = item.icon;
	const props = {
		href: item.href,
		className: ITEM,
		"aria-current": current ? ("page" as const) : undefined,
	};
	const body = (
		<>
			<Icon aria-hidden="true" />
			{t(`admin.nav.${item.key}`)}
		</>
	);
	// Framework screens are plain HTML pages: a full load, not an Inertia visit.
	return item.inertia ? <Link {...props}>{body}</Link> : <a {...props}>{body}</a>;
}

function Brand(): ReactElement {
	const t = useT();
	return (
		<div className="flex h-14 shrink-0 items-center gap-2 border-b border-border px-4">
			<span className="text-[15px] font-semibold tracking-tight">{BRAND}</span>
			<span className="rounded-md bg-surface px-1.5 py-0.5 text-[11px] font-medium text-muted">
				{t("admin.title")}
			</span>
		</div>
	);
}

function SidebarNav(): ReactElement {
	const t = useT();
	const { url } = usePage();
	const current = activeHref(url.split("?")[0]);
	return (
		<>
			<Brand />
			<nav aria-label={t("admin.topbar.navLabel")} className="flex-1 overflow-y-auto p-3">
				{NAV.map((group) => (
					<div key={group.key} className="mb-5 last:mb-0">
						<h2 className="mb-1 px-2 text-xs font-medium text-muted">
							{t(`admin.groups.${group.key}`)}
						</h2>
						<ul className="grid gap-0.5">
							{group.items.map((item) => (
								<li key={item.href}>
									<NavLink item={item} current={item.href === current} />
								</li>
							))}
						</ul>
					</div>
				))}
			</nav>
			<div className="border-t border-border p-3">
				<a href="/" className={ITEM}>
					<ArrowLeft aria-hidden="true" />
					{t("admin.nav.backToSite")}
				</a>
			</div>
		</>
	);
}

function ThemeToggle(): ReactElement {
	const t = useT();
	return (
		<Button
			variant="ghost"
			size="icon"
			aria-label={t("admin.topbar.toggleTheme")}
			onClick={toggleTheme}
		>
			<Sun aria-hidden="true" className="hidden dark:block" />
			<Moon aria-hidden="true" className="dark:hidden" />
		</Button>
	);
}

function UserMenu(): ReactElement | null {
	const t = useT();
	const { auth } = usePage<SharedProps>().props;
	const user = auth?.user;
	if (!user) return null;
	const name = user.nickname ?? user.name ?? "";
	return (
		<DropdownMenu>
			<DropdownMenuTrigger
				aria-label={t("admin.topbar.account")}
				className="rounded-full outline-none focus-visible:ring-2 focus-visible:ring-primary"
			>
				<Avatar name={name} />
			</DropdownMenuTrigger>
			<DropdownMenuContent>
				<DropdownMenuLabel>{t("admin.topbar.signedInAs", { name })}</DropdownMenuLabel>
				<DropdownMenuSeparator />
				<DropdownMenuItem render={<a href="/" />}>
					<ArrowLeft aria-hidden="true" />
					{t("admin.nav.backToSite")}
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}

/** The admin shell: fixed sidebar on large screens, a sheet on small ones, a topbar with the theme switch. */
export default function AdminLayout({ children }: { children: ReactNode }): ReactElement {
	const t = useT();
	const [open, setOpen] = useState(false);
	return (
		<div className="min-h-screen bg-bg text-fg">
			<a
				href="#admin-main"
				className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[80] focus:rounded-md focus:bg-fg focus:px-3 focus:py-2 focus:text-bg"
			>
				{t("admin.topbar.skip")}
			</a>
			<aside className="fixed inset-y-0 left-0 z-30 hidden w-64 flex-col border-r border-border bg-bg lg:flex">
				<SidebarNav />
			</aside>
			<Sheet open={open} onOpenChange={setOpen} title={t("admin.topbar.navLabel")}>
				<SidebarNav />
			</Sheet>
			<div className={cn("flex min-h-screen flex-col lg:pl-64")}>
				<header className="sticky top-0 z-20 flex h-14 items-center gap-2 border-b border-border bg-bg px-4 lg:px-8">
					<Button
						variant="ghost"
						size="icon"
						className="lg:hidden"
						aria-label={t("admin.topbar.openMenu")}
						onClick={() => setOpen(true)}
					>
						<Menu aria-hidden="true" />
					</Button>
					<div className="flex-1" />
					<ThemeToggle />
					<UserMenu />
				</header>
				<main id="admin-main" className="mx-auto w-full max-w-7xl flex-1 px-4 py-6 lg:px-8">
					{children}
				</main>
			</div>
		</div>
	);
}
