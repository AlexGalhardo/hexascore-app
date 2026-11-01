import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react-native";
import { type ReactElement, useSyncExternalStore } from "react";
import { ApiError } from "../../../../packages/app-client/src/api";
import { t } from "../i18n";
import { createLocaleStore } from "../i18n/locale";
import { memoryPreferences } from "../platform/preferences";
import { LocaleProvider } from "../providers/root";
import type { ProfileApi } from "./me";
import { ProfileSettings } from "./Settings";

const ME = {
	id: "7",
	name: "Ana",
	email: "ana@example.com",
	locale: "en",
	profilePrivate: true,
	personalizationOptOut: false,
	deletionScheduledAt: null,
};
const DEVICES = [
	{ id: "t1", name: "HexaScore · Pixel 8", lastUsedAt: null, expiresAt: null, current: true },
	{ id: "t2", name: "HexaScore · Galaxy", lastUsedAt: null, expiresAt: null, current: false },
];

function fakeApi(overrides: { meUpdate?: jest.Mock } = {}) {
	const client = {
		meUpdate:
			overrides.meUpdate ??
			jest.fn(async (args: { body?: unknown }) => ({ ...ME, ...(args.body as object) })),
		tokensRevoke: jest.fn(async () => undefined),
	};
	const read = { meShow: jest.fn(async () => ME), tokensIndex: jest.fn(async () => DEVICES) };
	const api: ProfileApi = { read, write: (call) => call(client) };
	return { api, client, read };
}

type Extra = { onCurrentRevoked?: () => void; signedIn?: boolean };

function Harness({ api, extra }: { api: ProfileApi; extra: Extra }): ReactElement {
	const store = useStore();
	const locale = useSyncExternalStore(store.subscribe, store.get);
	return (
		<LocaleProvider locale={locale}>
			<ProfileSettings
				api={api}
				signedIn={extra.signedIn ?? true}
				locale={locale}
				onLocale={store.set}
				theme="system"
				onTheme={() => {}}
				crashReport={null}
				onSignOut={() => {}}
				onCurrentRevoked={extra.onCurrentRevoked ?? (() => {})}
				onSignIn={() => {}}
			/>
		</LocaleProvider>
	);
}

let localeStore = createLocaleStore("en");
const useStore = () => localeStore;

async function renderProfile(api: ProfileApi, extra: Extra = {}) {
	const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	await render(
		<QueryClientProvider client={queryClient}>
			<Harness api={api} extra={extra} />
		</QueryClientProvider>,
	);
}

beforeEach(() => {
	localeStore = createLocaleStore("en");
});

describe("profile tab", () => {
	test("shows the account name and e-mail from GET /me", async () => {
		const { api } = fakeApi();
		await renderProfile(api);
		expect(await screen.findByText("ana@example.com")).toBeTruthy();
		expect(screen.getByText("Ana")).toBeTruthy();
	});

	test("signed out: GET /me and GET /tokens are not called", async () => {
		const { api, read } = fakeApi();
		await renderProfile(api, { signedIn: false });
		expect(screen.getByRole("button", { name: "Log in" })).toBeTruthy();
		expect(read.meShow).not.toHaveBeenCalled();
		expect(read.tokensIndex).not.toHaveBeenCalled();
	});

	test("422 nickname-taken is shown inline under the nickname field", async () => {
		const meUpdate = jest.fn(async () => {
			throw new ApiError(422, { status: 422, errors: { username: ["nickname-taken"] } });
		});
		const { api } = fakeApi({ meUpdate });
		await renderProfile(api);
		await fireEvent.changeText(await screen.findByLabelText("Nickname"), "taken");
		await fireEvent.press(screen.getByRole("button", { name: "Save name" }));
		expect(await screen.findByText("This nickname is taken. Try another one.")).toBeTruthy();
		expect(meUpdate).toHaveBeenCalledWith({
			body: {
				name: "Ana",
				locale: "en",
				profilePrivate: true,
				personalizationOptOut: false,
				username: "taken",
			},
		});
	});

	test("changing the locale calls PATCH /me {locale} and re-renders without a restart", async () => {
		const { api, client } = fakeApi();
		await renderProfile(api);
		await fireEvent.press(await screen.findByRole("radio", { name: "Español" }));
		expect(client.meUpdate).toHaveBeenCalledWith({
			body: expect.objectContaining({ locale: "es", name: "Ana", profilePrivate: true }),
		});
		expect(
			await screen.findByRole("header", { name: t("es", "account.settings.title") }),
		).toBeTruthy();
		expect(localeStore.get()).toBe("es");
		expect(memoryPreferences.get("hexascore.locale")).toBe("es");
	});

	test("the personalisation opt-out toggle sends the full profile", async () => {
		const { api, client } = fakeApi();
		await renderProfile(api);
		await fireEvent(
			await screen.findByRole("switch", { name: "Turn off personalised recommendations" }),
			"valueChange",
			true,
		);
		expect(client.meUpdate).toHaveBeenCalledWith({
			body: { name: "Ana", locale: "en", profilePrivate: true, personalizationOptOut: true },
		});
	});

	test("revoking another device deletes its token and refetches the list", async () => {
		const onCurrentRevoked = jest.fn();
		const { api, client, read } = fakeApi();
		await renderProfile(api, { onCurrentRevoked });
		const row = await screen.findByTestId("device-t2");
		await fireEvent.press(within(row).getByRole("button", { name: "Log out" }));
		expect(client.tokensRevoke).toHaveBeenCalledWith({ params: { id: "t2" } });
		await waitFor(() => expect(read.tokensIndex).toHaveBeenCalledTimes(2));
		expect(onCurrentRevoked).not.toHaveBeenCalled();
	});

	test("revoking the current device logs out, even when the call fails", async () => {
		const onCurrentRevoked = jest.fn();
		const { api, client } = fakeApi();
		client.tokensRevoke.mockRejectedValueOnce(new TypeError("Network request failed"));
		await renderProfile(api, { onCurrentRevoked });
		const row = await screen.findByTestId("device-t1");
		expect(within(row).getByText("HexaScore · Pixel 8 · This device")).toBeTruthy();
		await fireEvent.press(within(row).getByRole("button", { name: "Log out" }));
		expect(client.tokensRevoke).toHaveBeenCalledWith({ params: { id: "t1" } });
		await waitFor(() => expect(onCurrentRevoked).toHaveBeenCalledTimes(1));
	});
});
