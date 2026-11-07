package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261107200000CreateGoogleNonces stores the nonces the Android app asks for before a Google sign-in. The server
// issues the nonce, the app puts it in the authorization request, Google copies it into the ID token and the token
// exchange consumes it once (a DELETE that affects one row). Only the SHA-256 of the nonce is kept, so a database
// read cannot replay a pending sign-in; expired rows are pruned whenever a new nonce is issued.
func m20261107200000CreateGoogleNonces() schema.Migration {
	return schema.Migration{
		ID: "20261107200000_create_google_nonces",
		Up: func(s *schema.Schema) error {
			return s.Create("google_nonces", func(t *schema.Table) {
				t.String("nonce_hash").Primary()
				t.Timestamp("expires_at").Index()
			})
		},
		Down: func(s *schema.Schema) error { return s.DropIfExists("google_nonces") },
	}
}
