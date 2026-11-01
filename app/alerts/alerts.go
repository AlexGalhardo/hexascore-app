// Package alerts emits the in-app alert lines of SDD 12 §7.5: an ERROR line with a stable alert=<code>
// key that Railway log search and log alerts match (D-35: no metrics). Lines carry counts and names, never
// ids, emails or raw paths.
package alerts

import (
	"context"
	"log/slog"
)

// Codes of SDD 12 §7.5.
const (
	A5  = "A5"  // SSR fallback on a public route
	A6  = "A6"  // same job kind failed two runs in a row
	A10 = "A10" // quota flush failing
)

// Error logs msg at ERROR with the alert key first (log nil uses slog.Default).
func Error(ctx context.Context, log *slog.Logger, code, msg string, args ...any) {
	if log == nil {
		log = slog.Default()
	}

	log.ErrorContext(ctx, msg, append([]any{"alert", code}, args...)...)
}
