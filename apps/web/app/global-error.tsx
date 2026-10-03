"use client";

// Last-resort boundary when the root layout itself fails; it cannot use i18n or app styles.
export default function GlobalError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
	return (
		<html lang="en">
			<body style={{ fontFamily: "system-ui, sans-serif", padding: 32 }}>
				<h1>Something went wrong</h1>
				<p>An unexpected error happened. Try again in a moment.</p>
				{error.digest ? (
					<p style={{ fontFamily: "monospace", fontSize: 12 }}>Reference: {error.digest}</p>
				) : null}
				<button type="button" onClick={reset}>
					Try again
				</button>
			</body>
		</html>
	);
}
