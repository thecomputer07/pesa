-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE foods (
  id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  source_id         BIGINT NOT NULL,
  source_record_id  BIGINT NULL,
  name_it           VARCHAR(255) NOT NULL,
  name_en           VARCHAR(255) NULL,
  synonyms          TEXT NULL,
  category          VARCHAR(64) NULL,
  state             TEXT NOT NULL DEFAULT 'raw'
                    CHECK (state IN ('raw','cooked','dried','dehydrated','frozen')),
  parent_food_id    BIGINT NULL,
  edible_pct        NUMERIC(5,2) NOT NULL DEFAULT 100.00,
  kcal              NUMERIC(7,2) NULL,
  protein_g         NUMERIC(6,2) NULL,
  fat_g             NUMERIC(6,2) NULL,
  carb_g            NUMERIC(6,2) NULL,
  sugar_g           NUMERIC(6,2) NULL,
  fiber_g           NUMERIC(6,2) NULL,
  water_g           NUMERIC(6,2) NULL,
  ash_g             NUMERIC(6,2) NULL,
  alcohol_g         NUMERIC(6,2) NULL,
  salt_g            NUMERIC(6,3) NULL,
  last_imported_at  TIMESTAMPTZ NULL,
  superseded_by     BIGINT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_foods_source        FOREIGN KEY (source_id)      REFERENCES food_sources(id),
  CONSTRAINT fk_foods_parent        FOREIGN KEY (parent_food_id) REFERENCES foods(id),
  CONSTRAINT fk_foods_superseded_by FOREIGN KEY (superseded_by)  REFERENCES foods(id)
);
CREATE INDEX idx_foods_source ON foods (source_id);
CREATE INDEX idx_foods_category ON foods (category);
CREATE INDEX idx_foods_state ON foods (state);
CREATE INDEX ft_foods_name ON foods USING GIN (
  (coalesce(name_it,'') || ' ' || coalesce(name_en,'') || ' ' || coalesce(synonyms,'')) gin_trgm_ops
);

CREATE TRIGGER foods_set_updated_at BEFORE UPDATE ON foods
  FOR EACH ROW EXECUTE FUNCTION pesa_touch_updated_at();

-- +goose Down
DROP TABLE foods;
