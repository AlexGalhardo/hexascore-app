// Shapes of the props the Go controllers send.

export interface User {
	id: number;
	name: string;
	email: string;
	email_verified_at: string | null;
	deletion_scheduled_at?: string | null;
	role?: string;
	nickname?: string | null;
}

export type Locale = "en" | "pt-BR" | "es";

export interface SeoMeta {
	title: string;
	description: string;
	canonical: string;
	alternates: Record<Locale | "x-default", string>;
	robots: string;
	og: { image?: string; type: string };
	jsonLd?: object[];
}

export interface SharedProps {
	appName: string;
	locale: Locale;
	locales: Locale[];
	errors: Record<string, string>;
	auth?: { user: User | null };
	seo?: SeoMeta;
	consent?: "all" | "essential";
	unreadNotifications?: number;
	[key: string]: unknown;
}
