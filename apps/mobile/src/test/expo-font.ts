// Jest stand-in for expo-font: icon fonts count as loaded, so @expo/vector-icons renders its glyph
// without resolving the bundled .ttf through the asset registry (which jest-expo cannot do).
export const isLoaded = (): boolean => true;
export const loadAsync = async (): Promise<void> => {};
