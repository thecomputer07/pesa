-- +goose Up
-- Schema from the subject-rights procedure (erasure log, Appendix A).
-- Do NOT modify columns without updating the procedure document in lockstep.
CREATE TABLE erasure_log (
  id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id_hash         CHAR(64) NOT NULL,
  email_hash           CHAR(64) NOT NULL,
  requested_at         TIMESTAMPTZ NOT NULL,
  executed_at          TIMESTAMPTZ NOT NULL,
  backup_tombstone_at  TIMESTAMPTZ NULL,
  channel              TEXT NOT NULL CHECK (channel IN ('email','self-service')),
  ticket_ref           VARCHAR(32) NOT NULL,
  right_exercised      VARCHAR(8) NOT NULL,
  notes_hash           CHAR(64) NULL
);
CREATE INDEX idx_executed_at ON erasure_log (executed_at);
CREATE INDEX idx_backup_tombstone_at ON erasure_log (backup_tombstone_at);

-- +goose Down
DROP TABLE erasure_log;
