-- +goose Up
CREATE TABLE imports_review (
  id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  source_id    BIGINT NOT NULL,
  external_id  VARCHAR(128) NOT NULL,
  raw_json     JSONB NOT NULL,
  reason       VARCHAR(255) NOT NULL,
  severity     TEXT NOT NULL DEFAULT 'warn' CHECK (severity IN ('warn','error')),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  resolved_at  TIMESTAMPTZ NULL,
  CONSTRAINT fk_imports_review_source FOREIGN KEY (source_id) REFERENCES food_sources(id)
);
CREATE INDEX idx_imports_review_unresolved ON imports_review (resolved_at, severity);

-- +goose Down
DROP TABLE imports_review;
