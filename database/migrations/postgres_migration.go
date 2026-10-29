package migrations

import "github.com/AlexGalhardo/vukem/schema"

// postgresMigration builds a migration whose statements run one `s.Raw` each on PostgreSQL and are skipped elsewhere (scaffold tests run on SQLite, VF-033).
func postgresMigration(id string, up, down []string) schema.Migration {
	run := func(stmts []string) func(*schema.Schema) error {
		return func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}

			for _, stmt := range stmts {
				if err := s.Raw(stmt); err != nil {
					return err
				}
			}

			return nil
		}
	}

	return schema.Migration{ID: id, Up: run(up), Down: run(down)}
}
