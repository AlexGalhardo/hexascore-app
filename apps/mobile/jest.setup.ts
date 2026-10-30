// Jest runs on Node. Expo SDK 58 does not ship JSC; the Android binary defines
// HermesInternal. Mirror that global so the engine assertion can run in CI.
const runtime = globalThis as { HermesInternal?: object };
if (runtime.HermesInternal === undefined) runtime.HermesInternal = {};
