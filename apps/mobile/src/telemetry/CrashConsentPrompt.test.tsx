import { fireEvent, render, screen } from "@testing-library/react-native";
import { CrashConsentPrompt } from "./CrashConsentPrompt";

test("the prompt offers both choices and reports the one taken", async () => {
	const onChoose = jest.fn();
	await render(<CrashConsentPrompt locale="en" onChoose={onChoose} />);
	await fireEvent.press(screen.getByText("Don't send"));
	await fireEvent.press(screen.getByText("Send reports"));
	expect(onChoose.mock.calls).toEqual([[false], [true]]);
});
