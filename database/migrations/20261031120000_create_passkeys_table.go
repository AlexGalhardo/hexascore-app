package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031120000CreatePasskeysTable stores WebAuthn credentials (SDD 03 §7.8).
// aaguid stays nullable: vukem v0.34.0 Credential has no AAGUID and the app does not parse attestation.
// user_handle is NULL for imported Better Auth rows until a signature pins it.
func m20261031120000CreatePasskeysTable() schema.Migration {
	return schema.Migration{
		ID: "20261031120000_create_passkeys_table",
		Up: func(s *schema.Schema) error {
			return s.Create("passkeys", func(t *schema.Table) {
				t.ID()
				t.ForeignID("user_id").Constrained("users").CascadeOnDelete()
				t.Binary("credential_id").Unique()
				t.Binary("public_key")
				t.BigInteger("algorithm")
				t.BigInteger("sign_count").Default(0)
				t.JSON("transports").Default("[]")
				t.String("aaguid").Nullable()
				t.Binary("user_handle").Nullable()
				t.Boolean("backup_eligible").Default(false)
				t.Boolean("backup_state").Default(false)
				t.String("name", 64)
				t.Timestamp("created_at")
				t.Timestamp("last_used_at").Nullable()
				t.Index("user_id")
			})
		},
		Down: func(s *schema.Schema) error { return s.DropIfExists("passkeys") },
	}
}
