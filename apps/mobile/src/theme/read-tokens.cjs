const fs = require("node:fs");
const path = require("node:path");

const source = fs.readFileSync(
	path.join(__dirname, "../../../../packages/tokens/tokens.ts"),
	"utf8",
);

function block(mode) {
	const start = source.indexOf(`\t${mode}: {`);
	if (start < 0) throw new Error(`missing ${mode} colour block in packages/tokens/tokens.ts`);
	const end = source.indexOf("\n\t}", start);
	const body = source.slice(start, end);
	const colors = {};
	for (const match of body.matchAll(/"([^"]+)": \{ oklch: "[^"]+", hex: "(#[0-9a-f]+)" \}/g)) {
		colors[match[1]] = match[2];
	}
	if (Object.keys(colors).length === 0) throw new Error(`no ${mode} hex colours parsed`);
	return colors;
}

function exportedArray(name) {
	const match = new RegExp(`export const ${name} = (\\[[^\\]]+\\])`).exec(source);
	if (!match) throw new Error(`missing export ${name}`);
	return JSON.parse(match[1]);
}

// The app wears the neutral (zinc, Vercel-like) skin of the web UI: neutral.json overrides the shared
// tokens it names; category accents, score bands, stars and danger stay as in packages/tokens.
const neutral = require("./neutral.json");

function palette(mode) {
	return { ...block(mode), ...neutral[mode] };
}

module.exports = {
	palette,
	spacing: exportedArray("spacing"),
	fontSize: exportedArray("fontSize"),
	radius: neutral.radius,
};
