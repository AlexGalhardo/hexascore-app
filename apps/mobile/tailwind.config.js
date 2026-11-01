const tokens = require("./src/theme/read-tokens.cjs");

const light = tokens.palette("light");
const dark = tokens.palette("dark");

/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./app/**/*.{ts,tsx}", "./src/**/*.{ts,tsx}"],
	presets: [require("nativewind/preset")],
	darkMode: "class",
	theme: {
		extend: {
			colors: Object.fromEntries(
				Object.keys(light).map((name) => [name, { DEFAULT: light[name], dark: dark[name] }]),
			),
			spacing: Object.fromEntries(tokens.spacing.map((step) => [String(step), `${step * 4}px`])),
			fontSize: Object.fromEntries(tokens.fontSize.map((size) => [String(size), `${size}px`])),
			borderRadius: Object.fromEntries(
				Object.entries(tokens.radius).map(([name, value]) => [name, `${value}px`]),
			),
			fontFamily: {
				display: ["Playfair Display"],
				text: ["Inter"],
			},
		},
	},
	plugins: [],
};
