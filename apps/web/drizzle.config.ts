import { defineConfig } from "drizzle-kit";

export default defineConfig({
	dialect: "postgresql",
	schema: "./lib/db/schema/index.ts",
	out: "./lib/db/migrations",
	casing: "snake_case",
	dbCredentials: { url: process.env.DATABASE_URL || "postgres://hexascore:hexascore@localhost:5432/hexascore" },
});
