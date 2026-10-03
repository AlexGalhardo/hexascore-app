-- Hand-edited: add the column nullable, backfill existing rows, then enforce NOT NULL (safe on non-empty tables).
ALTER TABLE "lists" ADD COLUMN "type" "item_type";--> statement-breakpoint
-- Type of the list's first item by position; empty lists fall back to 'game'.
UPDATE "lists" SET "type" = coalesce(
	(SELECT i."type" FROM "list_items" li JOIN "items" i ON i."id" = li."item_id"
	 WHERE li."list_id" = "lists"."id" ORDER BY li."position", li."item_id" LIMIT 1),
	'game'
);--> statement-breakpoint
ALTER TABLE "lists" ALTER COLUMN "type" SET NOT NULL;--> statement-breakpoint
DROP INDEX "lists_user_idx";--> statement-breakpoint
CREATE INDEX "lists_user_type_idx" ON "lists" USING btree ("user_id","type");
