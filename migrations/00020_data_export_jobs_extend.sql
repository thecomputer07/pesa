-- +goose Up
-- PRIV-01: extend data_export_jobs with download_token, bytes, error columns
-- and add 'ready' / 'expired' to the status CHECK.
ALTER TABLE data_export_jobs
  ADD COLUMN download_token CHAR(43) NULL,
  ADD COLUMN bytes BIGINT NULL,
  ADD COLUMN error TEXT NULL;

ALTER TABLE data_export_jobs DROP CONSTRAINT IF EXISTS data_export_jobs_status_check;
ALTER TABLE data_export_jobs ADD CONSTRAINT data_export_jobs_status_check
  CHECK (status IN ('pending','running','ready','done','failed','expired'));

CREATE INDEX idx_data_export_jobs_token ON data_export_jobs (download_token);

-- +goose Down
DROP INDEX IF EXISTS idx_data_export_jobs_token;
ALTER TABLE data_export_jobs DROP CONSTRAINT IF EXISTS data_export_jobs_status_check;
ALTER TABLE data_export_jobs ADD CONSTRAINT data_export_jobs_status_check
  CHECK (status IN ('pending','running','done','failed'));
ALTER TABLE data_export_jobs DROP COLUMN error;
ALTER TABLE data_export_jobs DROP COLUMN bytes;
ALTER TABLE data_export_jobs DROP COLUMN download_token;
