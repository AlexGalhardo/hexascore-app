import {
	Activity,
	BarChart3,
	CheckCircle2,
	Database,
	Flag,
	Gauge,
	GitMerge,
	History,
	Layers,
	LayoutDashboard,
	Lock,
	type LucideIcon,
	Mail,
	Package,
	PenLine,
	ScrollText,
	ShieldCheck,
	Star,
	ToggleRight,
	TriangleAlert,
	Users,
	Workflow,
	Zap,
} from "lucide-react";

export interface NavItem {
	/** Key under `admin.nav` in the language catalogues. */
	key: string;
	href: string;
	icon: LucideIcon;
	/** True for screens already served as Inertia pages; the others are full page loads (framework templates). */
	inertia?: boolean;
}

export interface NavGroup {
	/** Key under `admin.groups` in the language catalogues. */
	key: string;
	items: NavItem[];
}

export const ADMIN_HOME = "/admin/overview";

export const NAV: NavGroup[] = [
	{
		key: "overview",
		items: [{ key: "dashboard", href: ADMIN_HOME, icon: LayoutDashboard, inertia: true }],
	},
	{
		key: "catalogue",
		items: [
			{ key: "items", href: "/admin/r/items", icon: Package, inertia: true },
			{ key: "ingestion", href: "/admin/ingest", icon: Workflow, inertia: true },
			{ key: "sources", href: "/admin/ingestion", icon: Database, inertia: true },
			{ key: "matchReviews", href: "/admin/ingest/reviews", icon: Layers, inertia: true },
			{
				key: "fieldReviews",
				href: "/admin/ingest/field-reviews",
				icon: PenLine,
				inertia: true,
			},
			{
				key: "itemConflicts",
				href: "/admin/ingest/conflicts",
				icon: GitMerge,
				inertia: true,
			},
			{ key: "catalogueAudit", href: "/admin/catalog-audit", icon: History, inertia: true },
			{ key: "locks", href: "/admin/locks", icon: Lock, inertia: true },
		],
	},
	{
		key: "moderation",
		items: [
			{ key: "moderationQueue", href: "/admin/moderation", icon: ShieldCheck, inertia: true },
			{
				key: "moderationMetrics",
				href: "/admin/moderation/metrics",
				icon: BarChart3,
				inertia: true,
			},
			{ key: "reports", href: "/admin/r/reports", icon: Flag, inertia: true },
			{ key: "ratingFlags", href: "/admin/r/rating-flags", icon: Star, inertia: true },
			{ key: "scoreActions", href: "/admin/r/score-actions", icon: Gauge, inertia: true },
			{ key: "scoreBursts", href: "/admin/r/score-bursts", icon: Zap, inertia: true },
			{
				key: "formulaVersions",
				href: "/admin/r/score-formula-versions",
				icon: Gauge,
				inertia: true,
			},
		],
	},
	{
		key: "operations",
		items: [
			{ key: "jobRuns", href: "/admin/r/job-runs", icon: Activity, inertia: true },
			{
				key: "ingestFailures",
				href: "/admin/r/ingest-failures",
				icon: TriangleAlert,
				inertia: true,
			},
			{ key: "dqResults", href: "/admin/r/dq-results", icon: CheckCircle2, inertia: true },
			{ key: "featureFlags", href: "/admin/r/flags", icon: ToggleRight, inertia: true },
		],
	},
	{
		key: "people",
		items: [
			{ key: "users", href: "/admin/r/users", icon: Users, inertia: true },
			{
				key: "contactMessages",
				href: "/admin/r/contact-messages",
				icon: Mail,
				inertia: true,
			},
		],
	},
	{
		key: "system",
		items: [{ key: "auditLog", href: "/admin/r/audit-logs", icon: ScrollText, inertia: true }],
	},
];

/** The item whose href is the longest prefix of the path, so /admin/ingest/reviews does not light up /admin/ingest. */
export function activeHref(path: string): string | null {
	let best: string | null = null;
	for (const group of NAV) {
		for (const item of group.items) {
			const hit = path === item.href || path.startsWith(`${item.href}/`);
			if (hit && (best === null || item.href.length > best.length)) best = item.href;
		}
	}
	return best;
}
