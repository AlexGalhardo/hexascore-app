INSERT INTO data_sources (code, name, licence, url) VALUES ('wikipedia', 'Wikipedia', 'CC BY-SA 4.0', 'https://wikipedia.org') ON CONFLICT DO NOTHING;
UPDATE item_translations SET synopsis = 'Text', synopsis_source = 'wikipedia', synopsis_licence = 'CC BY-SA 4.0', synopsis_source_url = NULL WHERE item_id = '0192f1c6-0000-7000-8000-0000000000d1'
