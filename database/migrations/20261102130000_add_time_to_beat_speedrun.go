package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102130000AddTimeToBeatSpeedrun adds the fourth time-to-beat style, user-reported like the other three, with
// the same 30–600,000 minute bounds. A report may now hold only a speedrun time, so the "at least one style" check is
// recreated to include it; the old one is found by definition because it was created unnamed.
func m20261102130000AddTimeToBeatSpeedrun() schema.Migration {
	return postgresMigration("20261102130000_add_time_to_beat_speedrun",
		[]string{
			`ALTER TABLE time_to_beat ADD COLUMN speedrun_minutes int
  CONSTRAINT time_to_beat_speedrun_minutes_range CHECK (speedrun_minutes IS NULL OR speedrun_minutes BETWEEN 30 AND 600000)`,
			`DO $$
DECLARE c text;
BEGIN
  SELECT conname INTO c FROM pg_constraint
   WHERE conrelid = 'time_to_beat'::regclass AND contype = 'c' AND pg_get_constraintdef(oid) ILIKE '%coalesce%';
  IF c IS NULL THEN
    RAISE EXCEPTION 'time_to_beat: the at-least-one-style CHECK was not found';
  END IF;
  EXECUTE format('ALTER TABLE time_to_beat DROP CONSTRAINT %I', c);
END $$`,
			`ALTER TABLE time_to_beat ADD CONSTRAINT time_to_beat_has_style
  CHECK (coalesce(main_minutes, extras_minutes, completionist_minutes, speedrun_minutes) IS NOT NULL)`,
		},
		[]string{
			`DELETE FROM time_to_beat WHERE coalesce(main_minutes, extras_minutes, completionist_minutes) IS NULL`,
			`ALTER TABLE time_to_beat DROP CONSTRAINT IF EXISTS time_to_beat_has_style`,
			`ALTER TABLE time_to_beat ADD CHECK (coalesce(main_minutes, extras_minutes, completionist_minutes) IS NOT NULL)`,
			`ALTER TABLE time_to_beat DROP COLUMN IF EXISTS speedrun_minutes`,
		},
	)
}
