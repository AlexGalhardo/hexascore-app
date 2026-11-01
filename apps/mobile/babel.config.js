const nativewindPreset = require("nativewind/babel");

// Under bun's isolated linker the string plugin names inside the nativewind
// preset resolve from react-native-css-interop's own dir, where
// @babel/plugin-transform-react-jsx is unreachable. Resolve them from here and
// from babel-preset-expo (which depends on it) instead.
const searchPaths = [__dirname, require.resolve("babel-preset-expo")];
const resolvePlugin = (plugin) => {
	const [name, ...rest] = Array.isArray(plugin) ? plugin : [plugin];
	if (typeof name !== "string") return plugin;
	const resolved = require.resolve(name, { paths: searchPaths });
	return rest.length ? [resolved, ...rest] : resolved;
};

const nativewindResolved = (api) => {
	const preset = nativewindPreset(api);
	return { ...preset, plugins: preset.plugins.map(resolvePlugin) };
};

module.exports = (api) => {
	api.cache(true);
	return {
		presets: [["babel-preset-expo", { jsxImportSource: "nativewind" }], nativewindResolved],
	};
};
