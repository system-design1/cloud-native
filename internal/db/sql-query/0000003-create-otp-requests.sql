-- +migrate Up
CREATE TABLE IF NOT EXISTS otp_requests (
  id BIGSERIAL PRIMARY KEY,

  request_id TEXT NOT NULL,
  tenant_id BIGINT NOT NULL,
  phone TEXT NOT NULL,

  status TEXT NOT NULL,
  provider_name TEXT NOT NULL DEFAULT '',
  provider_response JSONB NOT NULL DEFAULT '{}'::jsonb,
  error_message TEXT,

  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  correlation_id TEXT,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_otp_requests_request_id
  ON otp_requests (request_id);

CREATE INDEX IF NOT EXISTS ix_otp_requests_tenant_created_at
  ON otp_requests (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_otp_requests_phone_created_at
  ON otp_requests (phone, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_otp_requests_status_created_at
  ON otp_requests (status, created_at DESC);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS trg_otp_requests_updated_at ON otp_requests;

CREATE TRIGGER trg_otp_requests_updated_at
BEFORE UPDATE ON otp_requests
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +migrate Down
DROP TRIGGER IF EXISTS trg_otp_requests_updated_at ON otp_requests;
DROP INDEX IF EXISTS ix_otp_requests_status_created_at;
DROP INDEX IF EXISTS ix_otp_requests_phone_created_at;
DROP INDEX IF EXISTS ix_otp_requests_tenant_created_at;
DROP INDEX IF EXISTS ux_otp_requests_request_id;
DROP TABLE IF EXISTS otp_requests;
