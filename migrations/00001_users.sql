-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION pesa_touch_updated_at() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION pesa_touch_server_updated_at() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  NEW.server_updated_at := clock_timestamp();
  RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TABLE users (
  id                       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  email                    VARCHAR(255) NOT NULL,
  password_hash            VARCHAR(255) NOT NULL,
  email_verified_at        TIMESTAMPTZ NULL,
  created_at               TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at               TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  locale                   TEXT NOT NULL DEFAULT 'it' CHECK (locale IN ('it','en')),
  privacy_consent_at       TIMESTAMPTZ NOT NULL,
  privacy_consent_version  VARCHAR(20) NOT NULL,
  birth_date               DATE NULL,
  status                   TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','restricted','deleted')),
  CONSTRAINT uniq_users_email UNIQUE (email)
);

CREATE TRIGGER users_set_updated_at BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION pesa_touch_updated_at();

-- +goose Down
DROP TABLE users;
DROP FUNCTION IF EXISTS pesa_touch_updated_at();
DROP FUNCTION IF EXISTS pesa_touch_server_updated_at();
