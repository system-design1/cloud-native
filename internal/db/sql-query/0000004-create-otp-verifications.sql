-- +migrate Up
CREATE TABLE IF NOT EXISTS otp_verifications (
  id BIGSERIAL PRIMARY KEY,

  request_id TEXT NOT NULL DEFAULT '',
  tenant_id BIGINT NOT NULL,
  phone TEXT NOT NULL,

  result TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  attempt_count INTEGER NOT NULL DEFAULT 0,

  correlation_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_otp_verifications_request_id
  ON otp_verifications (request_id);

CREATE INDEX IF NOT EXISTS ix_otp_verifications_tenant_created_at
  ON otp_verifications (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_otp_verifications_phone_created_at
  ON otp_verifications (phone, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_otp_verifications_result_created_at
  ON otp_verifications (result, created_at DESC);

-- +migrate Down
DROP INDEX IF EXISTS ix_otp_verifications_result_created_at;
DROP INDEX IF EXISTS ix_otp_verifications_phone_created_at;
DROP INDEX IF EXISTS ix_otp_verifications_tenant_created_at;
DROP INDEX IF EXISTS ix_otp_verifications_request_id;
DROP TABLE IF EXISTS otp_verifications;
