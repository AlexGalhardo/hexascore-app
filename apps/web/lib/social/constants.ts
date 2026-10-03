// Client-safe constants (no drizzle imports) shared by the social UI and the server-side validation.

// Copy of the `item_status_value` enum (lib/db/schema/social.ts); repository.ts asserts they match.
export const ITEM_STATUSES = ["want", "in_progress", "completed", "dropped"] as const;
export type ItemStatus = (typeof ITEM_STATUSES)[number];

export const REPORT_REASONS = ["spam", "offensive", "harassment", "impersonation", "other"] as const;
export const REPORT_DETAILS_MAX = 500;

export const LIST_LIMITS = { lists: 50, items: 500, title: 80, description: 500 } as const;
