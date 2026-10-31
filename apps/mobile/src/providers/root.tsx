import {
	createContext,
	type ReactElement,
	type ReactNode,
	useContext,
	useEffect,
	useState,
} from "react";
import type { Session, SessionState } from "../auth/session";
import type { AppLocale } from "../i18n/locale";

export type RootDeps = {
	session: Pick<Session, "state" | "subscribe" | "restore">;
	/** Starts the foreground token rotation; returns the stop function. */
	startRotation: () => () => void;
	upgradeRequired: () => boolean;
	onUpgradeRequired: (listener: () => void) => () => void;
};

export type RouteGuards = {
	app: boolean;
	auth: boolean;
	update: boolean;
	banned: boolean;
	deletion: boolean;
};

// Update required blocks everything else (FR-10-5); logout stays on that screen.
// Banned and deletion-scheduled accounts see only their own screen (SDD 10 §6.2).
export function routeGuards(state: SessionState, upgrade: boolean): RouteGuards {
	const none = { app: false, auth: false, update: false, banned: false, deletion: false };
	if (upgrade) return { ...none, update: true };
	switch (state.kind) {
		case "loading":
			return none;
		case "active":
			return { ...none, app: true };
		case "banned":
			return { ...none, banned: true };
		case "deletionScheduled":
			return { ...none, deletion: true };
		default:
			return { ...none, auth: true };
	}
}

export function useRoot(deps: RootDeps): { state: SessionState; upgrade: boolean } {
	const [state, setState] = useState<SessionState>(deps.session.state());
	const [upgrade, setUpgrade] = useState<boolean>(deps.upgradeRequired());

	useEffect(() => {
		const unsubscribe = deps.session.subscribe(setState);
		const stopUpgrade = deps.onUpgradeRequired(() => setUpgrade(true));
		// The 426 may have arrived between the first render and this effect.
		if (deps.upgradeRequired()) setUpgrade(true);
		void deps.session.restore();
		const stopRotation = deps.startRotation();
		return () => {
			unsubscribe();
			stopUpgrade();
			stopRotation();
		};
	}, [deps]);

	return { state, upgrade };
}

const LocaleContext = createContext<AppLocale>("en");

export function LocaleProvider({
	locale,
	children,
}: {
	locale: AppLocale;
	children: ReactNode;
}): ReactElement {
	return <LocaleContext.Provider value={locale}>{children}</LocaleContext.Provider>;
}

export function useLocale(): AppLocale {
	return useContext(LocaleContext);
}
