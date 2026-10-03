"use client";

import { Menu } from "@base-ui/react/menu";
import { useLocale } from "next-intl";
import { getPathname, Link } from "@/i18n/navigation";
import { authClient } from "@/lib/auth-client";

type Props = {
	initial: string;
	nickname: string | null;
	admin: boolean;
	labels: { account: string; profile: string; api: string; logout: string; feed: string; publicProfile: string };
};

const itemClass = "flex min-h-11 cursor-default items-center px-3 text-sm outline-none data-[highlighted]:bg-surface";

export function UserMenu({ initial, nickname, admin, labels }: Props) {
	const locale = useLocale();

	async function logout(): Promise<void> {
		await authClient.signOut().catch(() => null);
		window.location.assign(getPathname({ href: "/", locale }));
	}

	return (
		<Menu.Root>
			<Menu.Trigger
				aria-label={labels.account}
				className="grid size-11 place-items-center rounded-full bg-fg font-serif text-lg font-bold text-bg"
			>
				{initial}
			</Menu.Trigger>
			<Menu.Portal>
				<Menu.Positioner align="end" sideOffset={6} className="z-50">
					<Menu.Popup className="min-w-44 rounded-md border border-border bg-bg py-1 shadow-lg">
						{nickname ? (
							<Menu.LinkItem
								render={<Link href={`/u/${encodeURIComponent(nickname)}`} />}
								className={itemClass}
							>
								{labels.publicProfile}
							</Menu.LinkItem>
						) : null}
						<Menu.LinkItem render={<Link href="/feed" />} className={itemClass}>
							{labels.feed}
						</Menu.LinkItem>
						<Menu.LinkItem render={<Link href="/profile" />} className={itemClass}>
							{labels.profile}
						</Menu.LinkItem>
						<Menu.LinkItem render={<Link href="/profile/api" />} className={itemClass}>
							{labels.api}
						</Menu.LinkItem>
						{admin ? (
							// Plain anchor: /admin is EN-only and has no localized prefix.
							<Menu.LinkItem render={<a href="/admin" />} className={itemClass}>
								Admin
							</Menu.LinkItem>
						) : null}
						<Menu.Separator className="my-1 h-px bg-border" />
						<Menu.Item onClick={logout} className={itemClass}>
							{labels.logout}
						</Menu.Item>
					</Menu.Popup>
				</Menu.Positioner>
			</Menu.Portal>
		</Menu.Root>
	);
}
