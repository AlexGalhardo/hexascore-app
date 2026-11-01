INSERT INTO items (id, type, slug, wikidata_qid, release_date, release_precision, release_status, provenance) VALUES
  ('0192f1c6-0000-7000-8000-0000000000d1', 'game', 'a-game', 'Q1', '2020-01-01', 'day', 'released', '{"title": "wikidata"}'),
  ('0192f1c6-0000-7000-8000-0000000000d2', 'book', 'a-book', 'Q2', '2000-01-01', 'day', 'released', '{}');
INSERT INTO item_translations (item_id, locale, title, title_source) VALUES
  ('0192f1c6-0000-7000-8000-0000000000d1', 'en', 'A Game', 'wikidata'), ('0192f1c6-0000-7000-8000-0000000000d2', 'en', 'A Book', 'wikidata');
INSERT INTO editions (id, work_id, isbn13, title, language, source) VALUES
  ('0192f1c6-0000-7000-8000-0000000000f1', '0192f1c6-0000-7000-8000-0000000000d2', '9780306406157', 'A Book', 'en', 'wikidata');
INSERT INTO entities (id, kind, slug, wikidata_qid, default_name) VALUES
  ('0192f1c6-0000-7000-8000-0000000000e1', 'person', 'a-person', 'Q10', 'A Person'), ('0192f1c6-0000-7000-8000-0000000000e2', 'organization', 'a-studio', 'Q11', 'A Studio');
INSERT INTO item_entities (item_id, entity_id, role, source) VALUES
  ('0192f1c6-0000-7000-8000-0000000000d1', '0192f1c6-0000-7000-8000-0000000000e2', 'developer', 'wikidata'), ('0192f1c6-0000-7000-8000-0000000000d2', '0192f1c6-0000-7000-8000-0000000000e1', 'author', 'wikidata');
