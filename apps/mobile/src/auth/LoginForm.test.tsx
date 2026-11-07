import { fireEvent, render, screen, waitFor } from "@testing-library/react-native";
import { LoginForm, type LoginFormProps } from "./LoginForm";
import type { LoginOutcome } from "./login";

// Title and submit button share the label; the button is the second match.
const submitForm = async (): Promise<void> => {
	await fireEvent.press(screen.getAllByText("Log in")[1] as never);
};
const token = { token: "t", tokenId: "i", expiresAt: "x" };

async function setup(outcomes: LoginOutcome[]) {
	const submit = jest.fn(async () => outcomes.shift() as LoginOutcome);
	const props: LoginFormProps = {
		locale: "en",
		initialDeviceName: "HexaScore · Pixel",
		submit,
		onSignedIn: jest.fn(async () => {}),
		openWeb: jest.fn(),
	};
	await render(<LoginForm {...props} />);
	await fireEvent.changeText(screen.getByLabelText("Email or nickname"), "ana");
	await fireEvent.changeText(screen.getByLabelText("Password"), "pw");
	return { submit, props };
}

test("two-factor-required shows the code field and a recovery code is submitted", async () => {
	const { submit, props } = await setup([{ kind: "twoFactor" }, { kind: "ok", token }]);
	await submitForm();
	await waitFor(() => screen.getByLabelText("Code"));
	await fireEvent.press(screen.getByText("Use a backup code"));
	await fireEvent.changeText(screen.getByLabelText("Backup code"), " abcd-efgh ");
	await fireEvent.press(screen.getByText("Verify"));
	await waitFor(() => expect(props.onSignedIn).toHaveBeenCalledWith(token));
	expect(submit).toHaveBeenLastCalledWith({
		identifier: "ana",
		password: "pw",
		deviceName: "HexaScore · Pixel",
		recoveryCode: "abcd-efgh",
	});
});

test("428 opens the web challenge", async () => {
	const { props } = await setup([{ kind: "challenge" }]);
	await submitForm();
	await waitFor(() => expect(props.openWeb).toHaveBeenCalledWith("/login"));
});

test("an empty device name blocks submit", async () => {
	const { submit } = await setup([]);
	await fireEvent.changeText(screen.getByDisplayValue("HexaScore · Pixel"), "   ");
	await submitForm();
	expect(submit).not.toHaveBeenCalled();
});

test("errors render as translated text", async () => {
	await setup([{ kind: "error", key: "auth.errors.invalidCredentials", requestId: undefined }]);
	await submitForm();
	expect(await screen.findByText("Wrong email/nickname or password.")).toBeTruthy();
});

test("the device name field is labelled through the catalogues", async () => {
	await render(
		<LoginForm
			locale="pt-BR"
			initialDeviceName="HexaScore · Pixel"
			submit={jest.fn()}
			onSignedIn={jest.fn(async () => {})}
			openWeb={jest.fn()}
		/>,
	);
	expect(screen.getByLabelText("Nome do dispositivo")).toBeTruthy();
});

describe("Google sign-in and web links", () => {
	const base = (): LoginFormProps => ({
		locale: "en",
		initialDeviceName: "HexaScore · Pixel",
		submit: jest.fn(),
		onSignedIn: jest.fn(async () => {}),
		openWeb: jest.fn(),
	});

	test("no Google button without a configured Google client", async () => {
		await render(<LoginForm {...base()} />);
		expect(screen.queryByText("Sign in with Google")).toBeNull();
	});

	test("Google signs in with the ID token and the device name", async () => {
		const props = {
			...base(),
			google: {
				prompt: jest.fn(async () => "idtok"),
				submit: jest.fn(async () => ({ kind: "ok", token }) as LoginOutcome),
			},
		};
		await render(<LoginForm {...props} />);
		await fireEvent.press(screen.getByText("Sign in with Google"));
		await waitFor(() => expect(props.onSignedIn).toHaveBeenCalledWith(token));
		expect(props.google.submit).toHaveBeenCalledWith({
			idToken: "idtok",
			deviceName: "HexaScore · Pixel",
		});
	});

	test("a cancelled Google prompt does nothing", async () => {
		const props = {
			...base(),
			google: { prompt: jest.fn(async () => null), submit: jest.fn() },
		};
		await render(<LoginForm {...props} />);
		await fireEvent.press(screen.getByText("Sign in with Google"));
		await waitFor(() => expect(props.google.prompt).toHaveBeenCalled());
		expect(props.google.submit).not.toHaveBeenCalled();
	});

	test("Google two-factor keeps the ID token for the code step", async () => {
		const submit = jest
			.fn()
			.mockResolvedValueOnce({ kind: "twoFactor" })
			.mockResolvedValueOnce({ kind: "ok", token });
		const props = { ...base(), google: { prompt: jest.fn(async () => "idtok"), submit } };
		await render(<LoginForm {...props} />);
		await fireEvent.press(screen.getByText("Sign in with Google"));
		await waitFor(() => screen.getByLabelText("Code"));
		await fireEvent.changeText(screen.getByLabelText("Code"), "123456");
		await fireEvent.press(screen.getByText("Verify"));
		await waitFor(() =>
			expect(submit).toHaveBeenLastCalledWith({
				idToken: "idtok",
				deviceName: "HexaScore · Pixel",
				code: "123456",
			}),
		);
	});

	test("forgot password and create account are links that open the web", async () => {
		const props = base();
		await render(<LoginForm {...props} />);
		await fireEvent.press(screen.getByRole("link", { name: "Forgot your password?" }));
		await fireEvent.press(screen.getByRole("link", { name: "Create account" }));
		expect(props.openWeb).toHaveBeenCalledWith("/forgot-password");
		expect(props.openWeb).toHaveBeenCalledWith("/signup");
	});
});
