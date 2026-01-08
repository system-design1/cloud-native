-- 1) Optional: enum types
DO $$ BEGIN
  CREATE TYPE tenant_status AS ENUM ('active', 'inactive', 'suspended');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE sms_provider AS ENUM ('kavenegar', 'twilio', 'ghasedak', 'other');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

-- 2) Table
CREATE TABLE IF NOT EXISTS tenant_settings (
  id                BIGSERIAL PRIMARY KEY,

  tenant_code       TEXT NOT NULL,            -- unique lookup key (slug/code)
  name              TEXT NOT NULL,

  status            tenant_status NOT NULL DEFAULT 'active',
  otp_enabled       BOOLEAN NOT NULL DEFAULT TRUE,

  sms_provider      sms_provider NOT NULL DEFAULT 'other',
  sms_api_key       TEXT,                     -- in real prod: store encrypted or secret reference

  rate_limit_per_min INTEGER NOT NULL DEFAULT 60,

  signup_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at        TIMESTAMPTZ,

  timezone          TEXT NOT NULL DEFAULT 'UTC',
  metadata          JSONB NOT NULL DEFAULT '{}'::jsonb,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at        TIMESTAMPTZ
);

-- 3) Uniques
CREATE UNIQUE INDEX IF NOT EXISTS ux_tenant_settings_tenant_code
  ON tenant_settings (tenant_code);

-- 4) Indexes for common filters / high-load lookups
-- If you often do: WHERE tenant_code = $1 AND status='active' AND otp_enabled=true AND deleted_at IS NULL
CREATE INDEX IF NOT EXISTS ix_tenant_settings_lookup_active
  ON tenant_settings (tenant_code)
  WHERE status = 'active' AND otp_enabled = true AND deleted_at IS NULL;

-- If you sometimes scan for expiring tenants or validity checks:
CREATE INDEX IF NOT EXISTS ix_tenant_settings_expires_at
  ON tenant_settings (expires_at)
  WHERE deleted_at IS NULL;

-- If you filter by status frequently:
CREATE INDEX IF NOT EXISTS ix_tenant_settings_status
  ON tenant_settings (status)
  WHERE deleted_at IS NULL;

-- 5) Auto-update updated_at (optional but common)
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_tenant_settings_updated_at ON tenant_settings;

CREATE TRIGGER trg_tenant_settings_updated_at
BEFORE UPDATE ON tenant_settings
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
