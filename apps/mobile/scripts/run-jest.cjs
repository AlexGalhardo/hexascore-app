const { spawnSync } = require("node:child_process");
const path = require("node:path");

function modulesOf(id) {
	return path.resolve(path.dirname(require.resolve(id)), "..");
}

// Bun's isolated store keeps Babel plugins and css-interop next to the package
// that depends on them. @babel/core and the JSX runtime require() them from
// the app root, so Jest needs those folders on NODE_PATH before it starts.
const extras = [modulesOf("babel-preset-expo/package.json"), modulesOf("nativewind/package.json")];
const current = process.env.NODE_PATH;
process.env.NODE_PATH = [...extras, current].filter(Boolean).join(path.delimiter);

const jestBin = require.resolve("jest/bin/jest");
const result = spawnSync(process.execPath, [jestBin, ...process.argv.slice(2)], {
	stdio: "inherit",
	env: process.env,
});
process.exit(result.status === null ? 1 : result.status);
