import { act, render, screen } from "@testing-library/react-native";
import type { ReactElement } from "react";
import { Text } from "react-native";
import type { SessionState } from "../auth/session";
import { LocaleProvider, type RootDeps, routeGuards, useLocale, useRoot } from "./root";

function fakeDeps(upgradeAtStart = false) {
	let state: SessionState = { kind: "loading" };
	const listeners = new Set<(s: SessionState) => void>();
	let upgradeListener: (() => void) | null = null;
	const stopRotation = jest.fn();
	const deps: RootDeps = {
		session: {
			state: () => state,
			subscribe(listener) {
				listeners.add(listener);
				return () => listeners.delete(listener);
			},
			restore: jest.fn(async () => {
				state = { kind: "signedOut" };
				for (const l of listeners) l(state);
			}),
		},
		startRotation: jest.fn(() => stopRotation),
		upgradeRequired: () => upgradeAtStart,
		onUpgradeRequired(listener) {
			upgradeListener = listener;
			return () => {
				upgradeListener = null;
			};
		},
	};
	return { deps, stopRotation, fireUpgrade: () => upgradeListener?.() };
}

function Probe({ deps }: { deps: RootDeps }): ReactElement {
	const { state, upgrade } = useRoot(deps);
	return <Text>{`${state.kind}:${upgrade}`}</Text>;
}

function ShowLocale(): ReactElement {
	return <Text>{useLocale()}</Text>;
}

test("restores the session, starts rotation and stops it on unmount", async () => {
	const { deps, stopRotation } = fakeDeps();
	const view = await render(<Probe deps={deps} />);
	expect(deps.session.restore).toHaveBeenCalledTimes(1);
	expect(deps.startRotation).toHaveBeenCalledTimes(1);
	expect(screen.getByText("signedOut:false")).toBeTruthy();
	await view.unmount();
	expect(stopRotation).toHaveBeenCalledTimes(1);
});

test("a 426 seen by the client switches to the update screen", async () => {
	const { deps, fireUpgrade } = fakeDeps();
	await render(<Probe deps={deps} />);
	await act(async () => fireUpgrade());
	expect(screen.getByText("signedOut:true")).toBeTruthy();
});

test("a 426 seen before the root mounted is not lost", async () => {
	const { deps } = fakeDeps(true);
	await render(<Probe deps={deps} />);
	expect(screen.getByText("signedOut:true")).toBeTruthy();
});

test("guards: update screen wins, then the session picks app or login", () => {
	const active: SessionState = { kind: "active", emailNotVerified: false };
	expect(routeGuards(active, true)).toEqual({ app: false, auth: false, update: true });
	expect(routeGuards({ kind: "signedOut" }, true)).toEqual({
		app: false,
		auth: false,
		update: true,
	});
	expect(routeGuards(active, false)).toEqual({ app: true, auth: false, update: false });
	expect(routeGuards({ kind: "signedOut" }, false)).toEqual({
		app: false,
		auth: true,
		update: false,
	});
	expect(routeGuards({ kind: "banned", reason: null, expiresAt: null }, false).auth).toBe(true);
});

test("the locale provider feeds useLocale", async () => {
	await render(
		<LocaleProvider locale="pt-BR">
			<ShowLocale />
		</LocaleProvider>,
	);
	expect(screen.getByText("pt-BR")).toBeTruthy();
});
