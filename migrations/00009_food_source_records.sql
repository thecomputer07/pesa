-- +goose Up
CREATE TABLE food_source_records (
  id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  source_id     BIGINT NOT NULL,
  external_id   VARCHAR(128) NOT NULL,
  raw_json      JSONB NOT NULL,
  fetched_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  content_hash  CHAR(64) NOT NULL,
  CONSTRAINT fk_fsr_source FOREIGN KEY (source_id) REFERENCES food_sources(id),
  CONSTRAINT uniq_fsr_source_external UNIQUE (source_id, external_id)
);
CREATE INDEX idx_fsr_hash ON food_source_records (content_hash);

ALTER TABLE foods
  ADD CONSTRAINT fk_foods_source_record FOREIGN KEY (source_record_id)
  REFERENCES food_source_records(id);

-- +goose Down
ALTER TABLE foods DROP CONSTRAINT IF EXISTS fk_foods_source_record;
DROP TABLE food_source_records;
