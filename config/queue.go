package config

// DefaultQueueLimits caps ingestion at 2 simultaneous jobs whatever QUEUE_CONCURRENCY is (SDD 06 NFR-06-2).
const DefaultQueueLimits = "ingest=2"

// WithQueueLimits wraps getenv so that QUEUE_LIMITS, vukem's per-queue running-job cap, falls back to
// DefaultQueueLimits when the operator leaves it unset or empty. A deployment that forgets the variable
// keeps the cap; any non-empty value replaces it.
func WithQueueLimits(getenv func(string) (string, bool)) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := getenv(k)
		if k == "QUEUE_LIMITS" && v == "" {
			return DefaultQueueLimits, true
		}

		return v, ok
	}
}
