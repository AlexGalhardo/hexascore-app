"use client";

import { Menu } from "@base-ui/react/menu";
import { useLocale } from "next-intl";
import { Link, usePathname } from "@/i18n/navigation";
import { type Locale, routing } from "@/i18n/routing";
import { NavIcon } from "./nav-icons";

// The menu shows codes only; native names stay as accessible labels (a language is named in itself).
const NAMES: Record<Locale, string> = { en: "English", "pt-BR": "Português", es: "Español" };
const SHORT: Record<Locale, string> = { en: "EN", "pt-BR": "PT-BR", es: "ES" };

const itemClass =
	"flex min-h-11 cursor-default items-center justify-between gap-6 px-3 text-sm outline-none data-[highlighted]:bg-surface";

export function LocaleMenu({ label }: { label: string }) {
	const current = useLocale() as Locale;
	// ponytail: the query string is dropped on switch; read window.location.search here if filters must survive.
	const pathname = usePathname();

	return (
		<Menu.Root>
			<Menu.Trigger
				aria-label={`${label}: ${NAMES[current]}`}
				className="flex min-h-11 items-center gap-1.5 rounded-md border border-border px-2.5 text-sm hover:bg-surface"
			>
				<NavIcon name="language" />
				<span>{SHORT[current]}</span>
			</Menu.Trigger>
			<Menu.Portal>
				<Menu.Positioner align="end" sideOffset={6} className="z-50">
					<Menu.Popup className="min-w-28 rounded-md border border-border bg-bg py-1 shadow-lg">
						{routing.locales.map((locale) => (
							<Menu.LinkItem
								key={locale}
								render={<Link href={pathname} locale={locale} lang={locale} />}
								aria-current={locale === current ? "true" : undefined}
								aria-label={NAMES[locale]}
								title={NAMES[locale]}
								className={itemClass}
							>
								{SHORT[locale]}
								{locale === current && <span aria-hidden="true">✓</span>}
							</Menu.LinkItem>
						))}
					</Menu.Popup>
				</Menu.Positioner>
			</Menu.Portal>
		</Menu.Root>
	);
}
