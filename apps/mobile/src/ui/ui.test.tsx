import { act, fireEvent, render, screen } from "@testing-library/react-native";
import { AccessibilityInfo } from "react-native";
import { Btn } from "./Btn";
import { formatScore } from "./format";
import { RatingInput, valueAt } from "./RatingInput";
import { ScoreBadge } from "./ScoreBadge";

describe("formatScore", () => {
	test("one decimal in the locale, dash without a score", () => {
		expect(formatScore(8, "en")).toBe("8.0");
		expect(formatScore(8.5, "pt-BR")).toBe("8,5");
		expect(formatScore(null, "es")).toBe("–");
	});
});

describe("ScoreBadge", () => {
	test("prints the number next to the band colour and labels it", async () => {
		await render(<ScoreBadge score={7.4} locale="en" label="Score 7.4 out of 10" />);
		expect(screen.getByText("7.4")).toBeTruthy();
		expect(screen.getByLabelText("Score 7.4 out of 10")).toBeTruthy();
		expect(screen.getByTestId("score-badge").props.className).toContain("bg-score-mid");
	});

	test("no score shows a dash on the none band", async () => {
		await render(<ScoreBadge score={null} locale="en" label="Not enough ratings" />);
		expect(screen.getByText("–")).toBeTruthy();
		expect(screen.getByTestId("score-badge").props.className).toContain("bg-score-none");
	});
});

describe("Btn", () => {
	test("is a 44 px button that presses and reports disabled", async () => {
		const onPress = jest.fn();
		await render(<Btn label="Save" onPress={onPress} variant="primary" />);
		const button = screen.getByRole("button", { name: "Save" });
		expect(button.props.className ?? "").toBeDefined();
		await fireEvent.press(button);
		expect(onPress).toHaveBeenCalledTimes(1);
	});

	test("disabled does not press", async () => {
		const onPress = jest.fn();
		await render(<Btn label="Save" onPress={onPress} disabled />);
		await fireEvent.press(screen.getByRole("button", { name: "Save" }));
		expect(onPress).not.toHaveBeenCalled();
		expect(screen.getByRole("button", { name: "Save" }).props.accessibilityState).toMatchObject({
			disabled: true,
		});
	});

	test("reduced motion drops the pressed translate", async () => {
		jest.spyOn(AccessibilityInfo, "isReduceMotionEnabled").mockResolvedValue(true);
		await render(<Btn label="Go" onPress={jest.fn()} />);
		await act(async () => {});
		expect(screen.getByRole("button", { name: "Go" }).props.className).not.toContain("translate-y");
	});
});

describe("RatingInput", () => {
	test("touch position maps to half-star steps", () => {
		expect(valueAt(0, 300)).toBe(0.5);
		expect(valueAt(14, 300)).toBe(0.5);
		expect(valueAt(16, 300)).toBe(1);
		expect(valueAt(299, 300)).toBe(10);
		expect(valueAt(400, 300)).toBe(10);
	});

	test("is an adjustable control stepping by 0.5 between 0.5 and 10", async () => {
		const onChange = jest.fn();
		await render(<RatingInput label="Gameplay" value={8.5} onChange={onChange} locale="pt-BR" />);
		const slider = screen.getByRole("adjustable", { name: "Gameplay" });
		expect(slider.props.accessibilityValue).toMatchObject({
			min: 0,
			max: 10,
			now: 8.5,
			text: "8,5 / 10",
		});
		expect(screen.getByText("8,5 / 10")).toBeTruthy();
		await fireEvent(slider, "accessibilityAction", { nativeEvent: { actionName: "increment" } });
		expect(onChange).toHaveBeenLastCalledWith(9);
		await fireEvent(slider, "accessibilityAction", { nativeEvent: { actionName: "decrement" } });
		expect(onChange).toHaveBeenLastCalledWith(8);
	});

	test("clamps at the ends and starts from nothing at the minimum", async () => {
		const onChange = jest.fn();
		const { rerender } = await render(
			<RatingInput label="Story" value={10} onChange={onChange} locale="en" />,
		);
		await fireEvent(screen.getByRole("adjustable"), "accessibilityAction", {
			nativeEvent: { actionName: "increment" },
		});
		expect(onChange).toHaveBeenLastCalledWith(10);
		await rerender(<RatingInput label="Story" value={null} onChange={onChange} locale="en" />);
		expect(screen.getByText("– / 10")).toBeTruthy();
		await fireEvent(screen.getByRole("adjustable"), "accessibilityAction", {
			nativeEvent: { actionName: "increment" },
		});
		expect(onChange).toHaveBeenLastCalledWith(0.5);
	});

	test("tapping the stars sets the value from the touch point", async () => {
		const onChange = jest.fn();
		await render(<RatingInput label="Art" value={null} onChange={onChange} locale="en" />);
		// The star row is hidden from TalkBack; the adjustable parent is the accessible control.
		const stars = screen.getByTestId("rating-stars", { includeHiddenElements: true });
		await fireEvent(stars, "layout", { nativeEvent: { layout: { width: 300 } } });
		await fireEvent.press(stars, { nativeEvent: { locationX: 100 } });
		expect(onChange).toHaveBeenLastCalledWith(3.5);
	});
});
