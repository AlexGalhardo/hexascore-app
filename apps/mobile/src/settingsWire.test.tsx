import { fireEvent, render, screen } from "@testing-library/react-native";
import Profile from "../app/(tabs)/profile";
import { memoryPreferences } from "./platform/preferences";
import { LocaleProvider } from "./providers/root";
import { CrashReportSwitch } from "./telemetry/CrashReportSwitch";
import { readThemePreference } from "./theme";

const mockConsent = { value: "unset" as "unset" | "granted" | "denied" };
const mockSetConsent = jest.fn(async (granted: boolean) => {
	mockConsent.value = granted ? "granted" : "denied";
});
jest.mock("./telemetry/runtime", () => ({
	crashConsent: () => mockConsent.value,
	setCrashConsent: (granted: boolean) => mockSetConsent(granted),
}));

const mockNavigate = jest.fn();
jest.mock("expo-router", () => ({ router: { navigate: (href: string) => mockNavigate(href) } }));

beforeEach(() => {
	mockConsent.value = "unset";
	mockSetConsent.mockClear();
	mockNavigate.mockClear();
});

describe("profile tab settings", () => {
	test("renders the v1 settings page in the active locale", async () => {
		await render(
			<LocaleProvider locale="pt-BR">
				<Profile />
			</LocaleProvider>,
		);
		expect(screen.getByRole("header", { name: "Configurações" })).toBeTruthy();
		expect(screen.getByRole("radio", { name: "Escuro" })).toBeTruthy();
	});

	test("choosing a theme mode applies and stores it", async () => {
		await render(<Profile />);
		await fireEvent.press(screen.getByRole("radio", { name: "Dark" }));
		expect(readThemePreference(memoryPreferences)).toBe("dark");
		expect(screen.getByRole("radio", { name: "Dark" }).props.accessibilityState).toMatchObject({
			checked: true,
		});
		await fireEvent.press(screen.getByRole("radio", { name: "System" }));
		expect(readThemePreference(memoryPreferences)).toBe("system");
	});

	test("signed out: log in opens the login route", async () => {
		await render(<Profile />);
		await fireEvent.press(screen.getByRole("button", { name: "Log in" }));
		expect(mockNavigate).toHaveBeenCalledWith("/login");
	});
});

describe("CrashReportSwitch", () => {
	test("off until granted; toggling writes the consent both ways", async () => {
		await render(<CrashReportSwitch locale="en" />);
		const toggle = screen.getByRole("switch", { name: "Send crash reports" });
		expect(toggle.props.value).toBe(false);
		await fireEvent(toggle, "valueChange", true);
		expect(mockSetConsent).toHaveBeenLastCalledWith(true);
		expect(screen.getByRole("switch", { name: "Send crash reports" }).props.value).toBe(true);
		await fireEvent(
			screen.getByRole("switch", { name: "Send crash reports" }),
			"valueChange",
			false,
		);
		expect(mockSetConsent).toHaveBeenLastCalledWith(false);
	});

	test("a granted consent shows on", async () => {
		mockConsent.value = "granted";
		await render(<CrashReportSwitch locale="en" />);
		expect(screen.getByRole("switch", { name: "Send crash reports" }).props.value).toBe(true);
	});
});
