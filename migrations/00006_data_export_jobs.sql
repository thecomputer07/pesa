-- +goose Up
CREATE TABLE data_export_jobs (
  id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id       BIGINT NOT NULL,
  status        TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','running','done','failed')),
  requested_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at  TIMESTAMPTZ NULL,
  file_path     VARCHAR(512) NULL,
  expires_at    TIMESTAMPTZ NULL,
  error_reason  VARCHAR(512) NULL,
  CONSTRAINT fk_data_export_jobs_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_data_export_jobs_user ON data_export_jobs (user_id);
CREATE INDEX idx_data_export_jobs_status ON data_export_jobs (status);

-- +goose Down
DROP TABLE data_export_jobs;
