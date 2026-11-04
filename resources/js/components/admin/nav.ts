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
			{ key: "items", href: "/admin/items", icon: Package },
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
			{ key: "catalogueAudit", href: "/admin/catalog-audit", icon: History },
			{ key: "locks", href: "/admin/locks", icon: Lock },
		],
	},
	{
		key: "moderation",
		items: [
			{ key: "moderationQueue", href: "/admin/moderation", icon: ShieldCheck },
			{ key: "moderationMetrics", href: "/admin/moderation/metrics", icon: BarChart3 },
			{ key: "reports", href: "/admin/reports", icon: Flag },
			{ key: "ratingFlags", href: "/admin/rating-flags", icon: Star },
			{ key: "scoreActions", href: "/admin/score-actions", icon: Gauge },
			{ key: "scoreBursts", href: "/admin/score-bursts", icon: Zap },
			{ key: "formulaVersions", href: "/admin/score-formula-versions", icon: Gauge },
		],
	},
	{
		key: "operations",
		items: [
			{ key: "jobRuns", href: "/admin/job-runs", icon: Activity },
			{ key: "ingestFailures", href: "/admin/ingest-failures", icon: TriangleAlert },
			{ key: "dqResults", href: "/admin/dq-results", icon: CheckCircle2 },
			{ key: "featureFlags", href: "/admin/_flags", icon: ToggleRight },
		],
	},
	{
		key: "people",
		items: [
			{ key: "users", href: "/admin/users", icon: Users },
			{ key: "contactMessages", href: "/admin/contact-messages", icon: Mail },
		],
	},
	{
		key: "system",
		items: [{ key: "auditLog", href: "/admin/_audit", icon: ScrollText }],
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
