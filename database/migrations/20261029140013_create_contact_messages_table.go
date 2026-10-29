package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140013CreateContactMessagesTable is SDD 14 §5 with users.id as bigint (the SDD's text is superseded by SDD 03): the contact form inbox, legacy parity. `handled` is toggled in the admin panel.
func m20261029140013CreateContactMessagesTable() schema.Migration {
	return postgresMigration("20261029140013_create_contact_messages_table",
		[]string{
			`CREATE TABLE contact_messages (
  id uuid PRIMARY KEY,
  user_id bigint REFERENCES users(id) ON DELETE SET NULL,
  name text NOT NULL,
  email text NOT NULL,
  subject text NOT NULL CHECK (subject IN ('question','feedback','bug')),
  message text NOT NULL CHECK (char_length(message) BETWEEN 32 AND 512),
  handled boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX contact_messages_handled_idx ON contact_messages (handled, created_at)`,
		},
		[]string{`DROP TABLE IF EXISTS contact_messages`},
	)
}
