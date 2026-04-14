-- +goose Up
-- PRIV-04: tombstone catalog consumed by the next pg_dump run so that
-- restored backups do not resurrect deleted users.
CREATE TABLE tombstones (
  id                  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id_pseudonym   CHAR(64) NOT NULL,
  email_hash          CHAR(64) NOT NULL,
  erased_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  applied_to_backups  TEXT NULL,
  last_attempted_at   TIMESTAMPTZ NULL
);
CREATE INDEX idx_tombstones_pseudo ON tombstones (user_id_pseudonym);
CREATE INDEX idx_tombstones_pending ON tombstones (erased_at);

-- +goose Down
DROP TABLE tombstones;
