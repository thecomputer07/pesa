-- +goose Up
CREATE TABLE password_resets (
  token         CHAR(64) NOT NULL PRIMARY KEY,
  user_id       BIGINT NOT NULL,
  expires_at    TIMESTAMPTZ NOT NULL,
  used_at       TIMESTAMPTZ NULL,
  requested_ip  VARCHAR(45) NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_password_resets_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_password_resets_user ON password_resets (user_id);
CREATE INDEX idx_password_resets_expires ON password_resets (expires_at);

-- +goose Down
DROP TABLE password_resets;
