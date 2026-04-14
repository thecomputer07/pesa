-- +goose Up
CREATE TABLE email_verifications (
  token       CHAR(64) NOT NULL PRIMARY KEY,
  user_id     BIGINT NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_email_verifications_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_email_verifications_user ON email_verifications (user_id);
CREATE INDEX idx_email_verifications_expires ON email_verifications (expires_at);

-- +goose Down
DROP TABLE email_verifications;
