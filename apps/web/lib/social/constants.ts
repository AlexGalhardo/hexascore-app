// Client-safe constants (no drizzle imports) shared by the social UI and the server-side validation.

// Copy of the `item_status_value` enum (lib/db/schema/social.ts); repository.ts asserts they match.
export const ITEM_STATUSES = ["want", "in_progress", "completed", "dropped"] as const;
export type ItemStatus = (typeof ITEM_STATUSES)[number];

export const REPORT_REASONS = ["spam", "offensive", "harassment", "impersonation", "other"] as const;
export const REPORT_DETAILS_MAX = 500;

export const LIST_LIMITS = { lists: 50, items: 500, title: 80, description: 500 } as const;

// Time to beat (games): hours in 0.5 steps, stored as minutes (CHECK `time_to_beat_range` mirrors the bounds).
export const TIME_TO_BEAT_STYLES = ["main", "extras", "completionist"] as const;
export type TimeToBeatStyle = (typeof TIME_TO_BEAT_STYLES)[number];
export const TIME_TO_BEAT_HOURS = { min: 0.5, max: 10_000, step: 0.5 } as const;
/** Medians are hidden (null) below this many reports per style, like MIN_RATINGS_FOR_SCORE for scores. */
export const MIN_REPORTS_FOR_TIME_TO_BEAT = 3;
