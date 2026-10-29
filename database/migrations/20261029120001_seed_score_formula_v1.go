package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029120001SeedScoreFormulaV1 records formula v1 with the provisional
// m_t of SDD 02 §7.1 (game 10, movie 10, series 8, book 8). Tuning on the
// first staging copy of production data is a P2.F task (open issue 02-2);
// a change is a new version row, never an edit of this one.
func m20261029120001SeedScoreFormulaV1() schema.Migration {
	return schema.Migration{
		ID: "20261029120001_seed_score_formula_v1",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`INSERT INTO score_formula_versions
  (version, m_game, m_movie, m_series, m_book, params, activated_at, changelog)
VALUES (1, 10, 10, 8, 8,
  '{"age_weights":{"lt7":0.25,"lt30":0.5,"lt90":0.75,"else":1},"status_weights":{"none_or_want":0.7,"in_progress":0.85,"dropped":0.85,"completed":1,"completed_game_without_ttb":0.9},"min_visible":5,"prior_min_items":20,"prior_default":60,"confidence":{"medium":{"n":15,"se":4.0},"high":{"n":50,"se":2.0}}}',
  now(), 'Formula v1 (SDD 02 §7.1); provisional m_t until tuned on staging data (02-2).')`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DELETE FROM score_formula_versions WHERE version = 1`)
		},
	}
}
