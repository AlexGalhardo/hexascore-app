// The app ships on Android only (toast.android.ts); this fallback keeps Jest and tsc platform-free.
export function showToast(_message: string): void {}
