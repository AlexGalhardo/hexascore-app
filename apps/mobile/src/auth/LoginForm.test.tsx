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
