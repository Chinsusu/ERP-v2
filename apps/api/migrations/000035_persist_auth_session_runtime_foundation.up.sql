BEGIN;

CREATE TABLE IF NOT EXISTS core.auth_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  session_ref text NOT NULL,
  user_ref text NOT NULL,
  email text NOT NULL,
  display_name text NOT NULL,
  role_code text NOT NULL,
  permissions jsonb NOT NULL DEFAULT '[]'::jsonb,
  access_token_hash text NOT NULL,
  refresh_token_hash text NOT NULL,
  access_expires_at timestamptz NOT NULL,
  refresh_expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  rotated_at timestamptz,
  last_seen_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_auth_sessions_ref UNIQUE (session_ref),
  CONSTRAINT uq_auth_sessions_access_token_hash UNIQUE (access_token_hash),
  CONSTRAINT uq_auth_sessions_refresh_token_hash UNIQUE (refresh_token_hash),
  CONSTRAINT ck_auth_sessions_required CHECK (
    nullif(btrim(session_ref), '') IS NOT NULL
    AND nullif(btrim(user_ref), '') IS NOT NULL
    AND nullif(btrim(email), '') IS NOT NULL
    AND nullif(btrim(display_name), '') IS NOT NULL
    AND nullif(btrim(role_code), '') IS NOT NULL
    AND nullif(btrim(access_token_hash), '') IS NOT NULL
    AND nullif(btrim(refresh_token_hash), '') IS NOT NULL
  ),
  CONSTRAINT ck_auth_sessions_token_hash_length CHECK (
    length(access_token_hash) = 64
    AND length(refresh_token_hash) = 64
  ),
  CONSTRAINT ck_auth_sessions_permissions_array CHECK (
    jsonb_typeof(permissions) = 'array'
  ),
  CONSTRAINT ck_auth_sessions_expiry_order CHECK (
    access_expires_at < refresh_expires_at
  ),
  CONSTRAINT ck_auth_sessions_version CHECK (version > 0)
);

CREATE INDEX IF NOT EXISTS ix_auth_sessions_org_email
  ON core.auth_sessions(org_id, lower(email), created_at DESC);

CREATE INDEX IF NOT EXISTS ix_auth_sessions_access_expiry
  ON core.auth_sessions(access_expires_at);

CREATE INDEX IF NOT EXISTS ix_auth_sessions_refresh_expiry
  ON core.auth_sessions(refresh_expires_at);

CREATE INDEX IF NOT EXISTS ix_auth_sessions_active_access
  ON core.auth_sessions(access_token_hash, access_expires_at)
  WHERE revoked_at IS NULL AND rotated_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_auth_sessions_active_refresh
  ON core.auth_sessions(refresh_token_hash, refresh_expires_at)
  WHERE revoked_at IS NULL AND rotated_at IS NULL;

CREATE TABLE IF NOT EXISTS core.auth_login_failures (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  email_normalized text NOT NULL,
  attempts integer NOT NULL DEFAULT 0,
  first_failed_at timestamptz,
  locked_until timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_auth_login_failures_org_email UNIQUE (org_id, email_normalized),
  CONSTRAINT ck_auth_login_failures_required CHECK (
    nullif(btrim(email_normalized), '') IS NOT NULL
  ),
  CONSTRAINT ck_auth_login_failures_attempts CHECK (attempts >= 0),
  CONSTRAINT ck_auth_login_failures_version CHECK (version > 0)
);

CREATE INDEX IF NOT EXISTS ix_auth_login_failures_locked_until
  ON core.auth_login_failures(locked_until)
  WHERE locked_until IS NOT NULL;

COMMIT;
