CREATE TABLE "time_to_beat" (
	"user_id" text NOT NULL,
	"item_id" uuid NOT NULL,
	"main_minutes" integer,
	"extras_minutes" integer,
	"completionist_minutes" integer,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL,
	CONSTRAINT "time_to_beat_user_id_item_id_pk" PRIMARY KEY("user_id","item_id"),
	CONSTRAINT "time_to_beat_range" CHECK ((main_minutes is null or main_minutes between 30 and 600000) and (extras_minutes is null or extras_minutes between 30 and 600000) and (completionist_minutes is null or completionist_minutes between 30 and 600000)),
	CONSTRAINT "time_to_beat_any" CHECK (coalesce(main_minutes, extras_minutes, completionist_minutes) is not null)
);
--> statement-breakpoint
ALTER TABLE "time_to_beat" ADD CONSTRAINT "time_to_beat_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "time_to_beat" ADD CONSTRAINT "time_to_beat_item_id_items_id_fk" FOREIGN KEY ("item_id") REFERENCES "public"."items"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "time_to_beat_item_idx" ON "time_to_beat" USING btree ("item_id");