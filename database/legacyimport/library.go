package legacyimport

import "time"

// librarySQL copies statuses and time-to-beat unchanged (SDD 02 §5).
var librarySQL = []userSQL{
	{"item_status", `INSERT INTO item_status (user_id, item_id, status, updated_at)
SELECT n.id, s.item_id, s.status::text::item_status_value, s.updated_at
FROM legacy_import.item_status s JOIN users n ON n.legacy_id = s.user_id
WHERE s.user_id > $1 AND s.user_id <= $2
ON CONFLICT (user_id, item_id) DO NOTHING`, false},
	{"time_to_beat", `INSERT INTO time_to_beat (user_id, item_id, main_minutes, extras_minutes, completionist_minutes,
  created_at, updated_at)
SELECT n.id, t.item_id, t.main_minutes, t.extras_minutes, t.completionist_minutes, t.created_at, t.updated_at
FROM legacy_import.time_to_beat t JOIN users n ON n.legacy_id = t.user_id
WHERE t.user_id > $1 AND t.user_id <= $2
ON CONFLICT (user_id, item_id) DO NOTHING`, false},
}

func libraryBatch(now func() time.Time) BatchFunc {
	return userBatches("library", librarySQL, nil, nil, now)
}
