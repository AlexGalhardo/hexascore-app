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
-- Billing tables keep the legacy shape (apps/web/lib/db/schema/billing.ts); the enums are text here.
CREATE TABLE legacy_import.api_keys (
  id uuid PRIMARY KEY,
  user_id text NOT NULL REFERENCES legacy_import."user"(id) ON DELETE CASCADE,
  name text NOT NULL,
  prefix text NOT NULL,
  hash text NOT NULL UNIQUE,
  last_used_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE legacy_import.api_usage_daily (
  user_id text NOT NULL REFERENCES legacy_import."user"(id) ON DELETE CASCADE,
  day date NOT NULL,
  count integer NOT NULL,
  PRIMARY KEY (user_id, day)
);
CREATE TABLE legacy_import.subscriptions (
  user_id text PRIMARY KEY REFERENCES legacy_import."user"(id) ON DELETE CASCADE,
  plan text NOT NULL,
  provider text NOT NULL,
  status text NOT NULL,
  current_period_end timestamptz NOT NULL,
  cancel_at_period_end boolean NOT NULL DEFAULT false,
  provider_customer_id text,
  provider_ref text,
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE legacy_import.stripe_customers (
  user_id text PRIMARY KEY REFERENCES legacy_import."user"(id) ON DELETE CASCADE,
  customer_id text NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE legacy_import.processed_events (id text PRIMARY KEY, provider text NOT NULL, processed_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE legacy_import.credit_ledger (
  id uuid PRIMARY KEY,
  user_id text NOT NULL REFERENCES legacy_import."user"(id) ON DELETE CASCADE,
  delta integer NOT NULL,
  reason text NOT NULL,
  ref text UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE legacy_import.pix_charges (
  id uuid PRIMARY KEY,
  user_id text NOT NULL REFERENCES legacy_import."user"(id) ON DELETE CASCADE,
  provider_ref text NOT NULL UNIQUE,
  purpose text NOT NULL,
  amount_cents integer NOT NULL,
  status text NOT NULL DEFAULT 'pending',
  br_code text,
  br_code_base64 text,
  expires_at timestamptz,
  paid_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

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
-- Legacy keys are sha256(full key) in hex; this one is a test key, never issued.
INSERT INTO legacy_import.api_keys (id, user_id, name, prefix, hash, created_at) VALUES
  ('01920000-0000-7000-8000-000000000001', 'u01', 'ci', 'hx_live_Ab3d',
   encode(sha256('hx_live_Ab3dEf6hIj9kLm2nOp5qRs8tUv1wXy4z'::bytea), 'hex'), '2025-02-01 00:00+00');
INSERT INTO legacy_import.api_usage_daily VALUES ('u01', '2025-03-01', 120), ('u01', '2025-03-02', 80);
-- u01 pays Pro through Stripe; u04 is back on free (no row is imported for free).
INSERT INTO legacy_import.subscriptions (user_id, plan, provider, status, current_period_end, provider_customer_id, provider_ref) VALUES
  ('u01', 'pro', 'stripe', 'active', '2099-01-01 00:00+00', 'cus_legacy1', 'sub_legacy1'),
  ('u04', 'free', 'stripe', 'canceled', '2025-01-01 00:00+00', NULL, NULL);
INSERT INTO legacy_import.stripe_customers (user_id, customer_id) VALUES ('u01', 'cus_legacy1');
INSERT INTO legacy_import.processed_events (id, provider) VALUES ('evt1', 'stripe'), ('evt2', 'stripe'), ('evt3', 'abacatepay');
-- u01: 25000 bought, 1200 spent; u05 overspent (balance below zero becomes 0).
INSERT INTO legacy_import.credit_ledger (id, user_id, delta, reason, ref) VALUES
  ('01920000-0000-7000-8000-0000000000c1', 'u01', 25000, 'purchase', 'cs_legacy1'),
  ('01920000-0000-7000-8000-0000000000c2', 'u01', -1200, 'usage', 'usage:u01:2025-03-01'),
  ('01920000-0000-7000-8000-0000000000c3', 'u05', 100, 'purchase', 'cs_legacy2'),
  ('01920000-0000-7000-8000-0000000000c4', 'u05', -150, 'usage', 'usage:u05:2025-03-01');
INSERT INTO legacy_import.pix_charges (id, user_id, provider_ref, purpose, amount_cents, status, paid_at) VALUES
  ('01920000-0000-7000-8000-0000000000d1', 'u05', 'pix_legacy1', 'credits', 1990, 'paid', '2025-03-01 00:00+00');
