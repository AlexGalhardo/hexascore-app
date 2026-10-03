import { sql } from "drizzle-orm";
import {
	boolean,
	check,
	index,
	integer,
	pgEnum,
	pgTable,
	primaryKey,
	smallint,
	text,
	timestamp,
	uniqueIndex,
	uuid,
} from "drizzle-orm/pg-core";
import { user } from "./auth";
import { items, itemType } from "./catalog";

export const ratings = pgTable(
	"ratings",
	{
		userId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		itemId: uuid()
			.notNull()
			.references(() => items.id, { onDelete: "cascade" }),
		c1: smallint().notNull(),
		c2: smallint().notNull(),
		c3: smallint().notNull(),
		c4: smallint().notNull(),
		c5: smallint().notNull(),
		c6: smallint().notNull(),
		overall: smallint().generatedAlwaysAs(sql`round((c1 + c2 + c3 + c4 + c5 + c6) / 6.0)`),
		createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
		updatedAt: timestamp({ withTimezone: true })
			.notNull()
			.defaultNow()
			.$onUpdate(() => new Date()),
	},
	(t) => [
		primaryKey({ columns: [t.userId, t.itemId] }),
		index("ratings_item_idx").on(t.itemId, t.createdAt),
		index("ratings_user_idx").on(t.userId, t.createdAt),
		check(
			"ratings_range",
			sql`c1 between 0 and 100 and c2 between 0 and 100 and c3 between 0 and 100 and c4 between 0 and 100 and c5 between 0 and 100 and c6 between 0 and 100`,
		),
	],
);

export const itemStatusValue = pgEnum("item_status_value", ["want", "in_progress", "completed", "dropped"]);

export const itemStatus = pgTable(
	"item_status",
	{
		userId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		itemId: uuid()
			.notNull()
			.references(() => items.id, { onDelete: "cascade" }),
		status: itemStatusValue().notNull(),
		updatedAt: timestamp({ withTimezone: true })
			.notNull()
			.defaultNow()
			.$onUpdate(() => new Date()),
	},
	(t) => [primaryKey({ columns: [t.userId, t.itemId] }), index("item_status_user_idx").on(t.userId, t.status)],
);

// Crowdsourced "time to beat" for games (one row per user and game, styles optional). Minutes, so the 0.5 h input
// step stays an integer. Only written while the user's status is `completed`; leaving it deletes the row.
export const timeToBeat = pgTable(
	"time_to_beat",
	{
		userId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		itemId: uuid()
			.notNull()
			.references(() => items.id, { onDelete: "cascade" }),
		mainMinutes: integer(),
		extrasMinutes: integer(),
		completionistMinutes: integer(),
		createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
		updatedAt: timestamp({ withTimezone: true })
			.notNull()
			.defaultNow()
			.$onUpdate(() => new Date()),
	},
	(t) => [
		primaryKey({ columns: [t.userId, t.itemId] }),
		index("time_to_beat_item_idx").on(t.itemId),
		check(
			"time_to_beat_range",
			sql`(main_minutes is null or main_minutes between 30 and 600000) and (extras_minutes is null or extras_minutes between 30 and 600000) and (completionist_minutes is null or completionist_minutes between 30 and 600000)`,
		),
		check("time_to_beat_any", sql`coalesce(main_minutes, extras_minutes, completionist_minutes) is not null`),
	],
);

export const follows = pgTable(
	"follows",
	{
		followerId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		followeeId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
	},
	(t) => [
		primaryKey({ columns: [t.followerId, t.followeeId] }),
		index("follows_followee_idx").on(t.followeeId),
		check("follows_not_self", sql`follower_id <> followee_id`),
	],
);

export const lists = pgTable(
	"lists",
	{
		id: uuid()
			.primaryKey()
			.$defaultFn(() => Bun.randomUUIDv7()),
		userId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		// A list holds items of a single type; the per-type cap (LIST_LIMITS.perType) is counted on (user_id, type).
		type: itemType().notNull(),
		title: text().notNull(),
		description: text(),
		isPublic: boolean().notNull().default(true),
		hidden: boolean().notNull().default(false),
		createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
		updatedAt: timestamp({ withTimezone: true })
			.notNull()
			.defaultNow()
			.$onUpdate(() => new Date()),
	},
	(t) => [index("lists_user_type_idx").on(t.userId, t.type)],
);

export const listItems = pgTable(
	"list_items",
	{
		listId: uuid()
			.notNull()
			.references(() => lists.id, { onDelete: "cascade" }),
		itemId: uuid()
			.notNull()
			.references(() => items.id, { onDelete: "cascade" }),
		position: integer().notNull(),
		note: text(),
	},
	(t) => [primaryKey({ columns: [t.listId, t.itemId] })],
);

export const reportTarget = pgEnum("report_target", ["list", "user"]);
export const reportStatus = pgEnum("report_status", ["open", "resolved", "dismissed"]);

export const reports = pgTable(
	"reports",
	{
		id: uuid()
			.primaryKey()
			.$defaultFn(() => Bun.randomUUIDv7()),
		reporterId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		targetType: reportTarget().notNull(),
		targetId: text().notNull(),
		reason: text().notNull(),
		details: text(),
		status: reportStatus().notNull().default("open"),
		createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
	},
	(t) => [
		index("reports_status_idx").on(t.status, t.createdAt),
		// One open report per reporter and target: repeated clicks never pile up in the moderation queue.
		uniqueIndex("reports_open_unique_idx").on(t.reporterId, t.targetType, t.targetId).where(sql`status = 'open'`),
	],
);

export const contactMessages = pgTable("contact_messages", {
	id: uuid()
		.primaryKey()
		.$defaultFn(() => Bun.randomUUIDv7()),
	userId: text().references(() => user.id, { onDelete: "set null" }),
	name: text().notNull(),
	email: text().notNull(),
	subject: text().notNull(),
	message: text().notNull(),
	handled: boolean().notNull().default(false),
	createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
});
