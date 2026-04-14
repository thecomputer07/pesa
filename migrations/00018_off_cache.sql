-- +goose Up
CREATE TABLE off_cache (
  ean          CHAR(14) NOT NULL PRIMARY KEY,
  payload_json JSONB NULL,
  status       TEXT NOT NULL CHECK (status IN ('hit','miss','error')),
  fetched_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  error_msg    VARCHAR(255) NULL
);
CREATE INDEX idx_off_cache_status_fetched ON off_cache (status, fetched_at);

-- +goose Down
DROP TABLE off_cache;
