import { fireEvent, render, screen } from "@testing-library/react-native";
import { BannedBody, DeletionScheduledBody } from "./Blocked";

describe("banned screen", () => {
	test("shows the reason and the end date, and leaves to the login", async () => {
		const onLeave = jest.fn();
		await render(
			<BannedBody
				locale="en"
				reason="Score manipulation"
				expiresAt="2026-12-01T12:00:00Z"
				onLeave={onLeave}
			/>,
		);
		expect(screen.getByText("Your account is suspended")).toBeTruthy();
		expect(screen.getByText("Reason: Score manipulation")).toBeTruthy();
		expect(screen.getByText("Until December 1, 2026")).toBeTruthy();
		await fireEvent.press(screen.getByText("Back to login"));
		expect(onLeave).toHaveBeenCalledTimes(1);
	});

	test("a ban without an end date says so, and a missing reason is not shown", async () => {
		await render(<BannedBody locale="pt-BR" reason={null} expiresAt={null} onLeave={() => {}} />);
		expect(screen.getByText("Sua conta está suspensa")).toBeTruthy();
		expect(screen.getByText("Sem data de término")).toBeTruthy();
		expect(screen.queryByText(/Motivo/)).toBeNull();
	});
});

describe("deletion-scheduled screen", () => {
	test("shows the date and sends the user to web /login to cancel", async () => {
		const openWeb = jest.fn();
		const onLogout = jest.fn();
		await render(
			<DeletionScheduledBody
				locale="es"
				scheduledAt="2026-12-01T12:00:00Z"
				openWeb={openWeb}
				onLogout={onLogout}
			/>,
		);
		expect(screen.getByText(/1 de diciembre de 2026/)).toBeTruthy();
		await fireEvent.press(screen.getByText("Cancelar eliminación"));
		expect(openWeb).toHaveBeenCalledWith("/login");
		await fireEvent.press(screen.getByText("Volver al inicio de sesión"));
		expect(onLogout).toHaveBeenCalledTimes(1);
	});
});
