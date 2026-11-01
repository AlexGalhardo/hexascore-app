const path = require("node:path");
const { getDefaultConfig } = require("expo/metro-config");
const { withNativeWind } = require("nativewind/metro");

const projectRoot = __dirname;
const workspaceRoot = path.resolve(projectRoot, "../..");
const config = getDefaultConfig(projectRoot);

config.watchFolders = [workspaceRoot];
config.resolver.nodeModulesPaths = [
	path.resolve(projectRoot, "node_modules"),
	path.resolve(workspaceRoot, "node_modules"),
];

const nativewindConfig = withNativeWind(config, { input: "./global.css" });
const baseResolve = nativewindConfig.resolver.resolveRequest;

// Bun's isolated linker links react-native-css-interop only next to nativewind, but
// jsxImportSource "nativewind" injects react-native-css-interop/jsx-runtime into every app
// file. Resolve that package from nativewind's directory.
const nativewindDir = path.dirname(
	require.resolve("nativewind/package.json", { paths: [projectRoot] }),
);
const cssInteropDir = path.dirname(
	require.resolve("react-native-css-interop/package.json", { paths: [nativewindDir] }),
);

nativewindConfig.resolver.resolveRequest = (context, moduleName, platform) => {
	const name = moduleName.startsWith("react-native-css-interop/")
		? path.join(cssInteropDir, moduleName.slice("react-native-css-interop/".length))
		: moduleName;
	return (baseResolve ?? context.resolveRequest)(context, name, platform);
};

module.exports = nativewindConfig;
