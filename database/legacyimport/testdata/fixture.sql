-- Synthetic legacy dump restored as schema legacy_import (SDD 11 §7.6 step 3).
-- Better Auth tables keep the legacy column names (legacy apps/web/lib/db/schema/auth.ts);
-- the other tables carry only what the runner reads today and grow with each stage task.
-- No real data: every address is under example.test; secrets are placeholders, except a01 and a02, hashes of a public test password.
DROP SCHEMA IF EXISTS legacy_import CASCADE;
CREATE SCHEMA legacy_import;

CREATE TABLE legacy_import."user" (
  id text PRIMARY KEY,
  name text NOT NULL,
  email text NOT NULL UNIQUE,
  email_verified boolean NOT NULL DEFAULT false,
  image text,
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp,
  username text UNIQUE,
  display_username text,
  two_factor_enabled boolean DEFAULT false,
  role text,
  banned boolean DEFAULT false,
  ban_reason text,
  ban_expires timestamp,
  profile_private boolean NOT NULL DEFAULT false,
  deletion_scheduled_at timestamp
);

CREATE TABLE legacy_import.account (
  id text PRIMARY KEY,
  account_id text NOT NULL,
  provider_id text NOT NULL,
  user_id text NOT NULL REFERENCES legacy_import."user"(id) ON DELETE CASCADE,
  access_token text,
  refresh_token text,
  id_token text,
  password text,
  created_at timestamp NOT NULL DEFAULT now()
);

CREATE TABLE legacy_import.two_factor (
  id text PRIMARY KEY,
  secret text NOT NULL,
  backup_codes text NOT NULL,
  user_id text NOT NULL REFERENCES legacy_import."user"(id) ON DELETE CASCADE,
  verified boolean DEFAULT true,
  failed_verification_count integer DEFAULT 0,
  locked_until timestamp
);

CREATE TABLE legacy_import.passkey (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES legacy_import."user"(id) ON DELETE CASCADE
);

CREATE TABLE legacy_import.items (id text PRIMARY KEY);
CREATE TABLE legacy_import.item_translations (item_id text NOT NULL, locale text NOT NULL, PRIMARY KEY (item_id, locale));
CREATE TABLE legacy_import.source_records (id text PRIMARY KEY);
CREATE TABLE legacy_import.affiliate_links (id text PRIMARY KEY);
CREATE TABLE legacy_import.ratings (id text PRIMARY KEY);
CREATE TABLE legacy_import.lists (id text PRIMARY KEY);
CREATE TABLE legacy_import.list_items (id text PRIMARY KEY);
CREATE TABLE legacy_import.reports (id text PRIMARY KEY);
CREATE TABLE legacy_import.contact_messages (id text PRIMARY KEY);
CREATE TABLE legacy_import.item_status (id text PRIMARY KEY);
CREATE TABLE legacy_import.time_to_beat (id text PRIMARY KEY);
CREATE TABLE legacy_import.api_keys (id text PRIMARY KEY);
CREATE TABLE legacy_import.api_usage_daily (id text PRIMARY KEY);
CREATE TABLE legacy_import.subscriptions (id text PRIMARY KEY);
CREATE TABLE legacy_import.stripe_customers (id text PRIMARY KEY);
CREATE TABLE legacy_import.processed_events (id text PRIMARY KEY);
CREATE TABLE legacy_import.credit_ledger (id text PRIMARY KEY);

-- u01 mixed-case email, u03 unverified and scheduled for deletion, u04 admin with
-- a live ban, u06 with an expired ban.
INSERT INTO legacy_import."user" (id, name, email, email_verified, created_at, updated_at, username, display_username,
  two_factor_enabled, role, banned, ban_reason, ban_expires, profile_private, deletion_scheduled_at) VALUES
  ('u01', 'Ana', 'Ana@Example.test', true, '2025-01-02 10:00', '2025-01-03 11:00', 'anaana', 'AnaAna', false, 'user', false, NULL, NULL, true, NULL),
  ('u02', 'Bruno', 'bruno@example.test', true, '2025-01-02 10:00', NULL, NULL, NULL, true, NULL, false, NULL, NULL, false, NULL),
  ('u03', 'Carla', 'carla@example.test', false, '2025-01-02 10:00', NULL, NULL, NULL, false, 'user', false, NULL, NULL, false, '2099-01-01 00:00'),
  ('u04', 'Davi', 'davi@example.test', true, '2025-01-02 10:00', NULL, NULL, NULL, false, 'user,admin', true, 'spam', '2099-01-01 00:00', false, NULL),
  ('u05', 'Eva', 'eva@example.test', true, '2025-01-02 10:00', NULL, NULL, NULL, false, 'user', false, NULL, NULL, false, NULL),
  ('u06', 'Fabio', 'fabio@example.test', true, '2025-01-02 10:00', NULL, NULL, NULL, false, 'user', true, 'old', '2020-01-01 00:00', false, NULL),
  ('u07', 'Gil', 'gil@example.test', true, '2025-01-02 10:00', NULL, NULL, NULL, false, 'user', false, NULL, NULL, false, NULL);

-- a01 and a02 hold a Better Auth 1.7.7 hash of "correct horse battery staple" (testdata/betterauth_hashes.json).
INSERT INTO legacy_import.account (id, account_id, provider_id, user_id, access_token, refresh_token, id_token, password) VALUES
  ('a01', 'u01', 'credential', 'u01', NULL, NULL, NULL, 'f784aeba634fcc3813d2fed8886d1f4f:beaf7dbabd716f288bf679caddface38ac825862263cef82ed40e0c93f9949269bf3d57e5986a080b8005361fb11dde09950ec22d09924efbf33bf3b5cdf7a16'),
  ('a02', 'u02', 'credential', 'u02', NULL, NULL, NULL, 'f784aeba634fcc3813d2fed8886d1f4f:beaf7dbabd716f288bf679caddface38ac825862263cef82ed40e0c93f9949269bf3d57e5986a080b8005361fb11dde09950ec22d09924efbf33bf3b5cdf7a16'),
  ('a03', 'g-1003', 'google', 'u03', 'placeholder-access', 'placeholder-refresh', 'placeholder-id', NULL),
  ('a04', 'u04', 'credential', 'u04', NULL, NULL, NULL, 'placeholder-scrypt-hash'),
  ('a05', 'g-1005', 'google', 'u05', 'placeholder-access', NULL, NULL, NULL),
  ('a06', 'u06', 'credential', 'u06', NULL, NULL, NULL, 'placeholder-scrypt-hash'),
  ('a16', 'g-1006', 'google', 'u06', NULL, NULL, NULL, NULL);

-- t04 never finished the TOTP setup (verified false).
INSERT INTO legacy_import.two_factor (id, secret, backup_codes, user_id, verified) VALUES
  ('t02', 'placeholder-encrypted-secret', 'placeholder-encrypted-codes', 'u02', true),
  ('t04', 'placeholder-encrypted-secret', 'placeholder-encrypted-codes', 'u04', false);

-- u07 signs in with a passkey only: counted by the pre-flight, not an abort.
INSERT INTO legacy_import.passkey (id, user_id) VALUES ('p07', 'u07');

INSERT INTO legacy_import.items SELECT 'i' || lpad(g::text, 3, '0') FROM generate_series(1, 12) g;
INSERT INTO legacy_import.item_translations SELECT 'i' || lpad(g::text, 3, '0'), l FROM generate_series(1, 12) g, unnest(ARRAY['en', 'pt-BR']) l;
INSERT INTO legacy_import.source_records SELECT 's' || g FROM generate_series(1, 12) g;
INSERT INTO legacy_import.affiliate_links VALUES ('al1'), ('al2');
INSERT INTO legacy_import.ratings SELECT 'r' || g FROM generate_series(1, 20) g;
INSERT INTO legacy_import.lists VALUES ('l1'), ('l2');
INSERT INTO legacy_import.list_items SELECT 'li' || g FROM generate_series(1, 5) g;
INSERT INTO legacy_import.reports VALUES ('rp1');
INSERT INTO legacy_import.contact_messages VALUES ('cm1');
INSERT INTO legacy_import.item_status SELECT 'st' || g FROM generate_series(1, 8) g;
INSERT INTO legacy_import.time_to_beat SELECT 'ttb' || g FROM generate_series(1, 3) g;
INSERT INTO legacy_import.api_keys VALUES ('k1');
INSERT INTO legacy_import.api_usage_daily VALUES ('ud1'), ('ud2');
INSERT INTO legacy_import.subscriptions VALUES ('sub1');
INSERT INTO legacy_import.stripe_customers VALUES ('cus1');
INSERT INTO legacy_import.processed_events VALUES ('evt1'), ('evt2'), ('evt3');
INSERT INTO legacy_import.credit_ledger VALUES ('cl1');
