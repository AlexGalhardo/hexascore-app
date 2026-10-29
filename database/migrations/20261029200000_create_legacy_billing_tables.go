package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029200000CreateLegacyBillingTables keeps the legacy credit ledger and Pix charges in their
// legacy shape (users.id as bigint) for the second release (SDD 05 §5, SDD 11 §11 issue 9). v1 reads
// only api_credit_balances; nothing writes these tables after the import.
func m20261029200000CreateLegacyBillingTables() schema.Migration {
	return postgresMigration("20261029200000_create_legacy_billing_tables",
		[]string{
			`CREATE TABLE credit_ledger (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  delta integer NOT NULL,
  reason text NOT NULL,
  ref text UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX credit_ledger_user_idx ON credit_ledger (user_id)`,
			`CREATE TABLE pix_charges (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider_ref text NOT NULL UNIQUE,
  purpose text NOT NULL,
  amount_cents integer NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid','expired','failed')),
  br_code text,
  br_code_base64 text,
  expires_at timestamptz,
  paid_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX pix_charges_user_idx ON pix_charges (user_id)`,
		},
		[]string{`DROP TABLE IF EXISTS pix_charges`, `DROP TABLE IF EXISTS credit_ledger`},
	)
}
