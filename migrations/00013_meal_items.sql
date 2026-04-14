-- +goose Up
-- Note: recipe_id FK is added in 00014 after recipes table exists (forward ref avoidance).
CREATE TABLE meal_items (
  id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  meal_id           BIGINT NOT NULL,
  client_id         CHAR(36) NOT NULL,
  food_id           BIGINT NULL,
  recipe_id         BIGINT NULL,
  name_snapshot     VARCHAR(255) NOT NULL,
  grams             NUMERIC(7,2) NOT NULL,
  kcal              NUMERIC(7,2) NULL,
  protein_g         NUMERIC(6,2) NULL,
  fat_g             NUMERIC(6,2) NULL,
  carb_g            NUMERIC(6,2) NULL,
  sugar_g           NUMERIC(6,2) NULL,
  fiber_g           NUMERIC(6,2) NULL,
  updated_at_ms     BIGINT NOT NULL,
  deleted_at_ms     BIGINT NULL,
  server_updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_meal_items_meal FOREIGN KEY (meal_id) REFERENCES meals(id) ON DELETE CASCADE,
  CONSTRAINT fk_meal_items_food FOREIGN KEY (food_id) REFERENCES foods(id),
  CONSTRAINT chk_meal_items_target CHECK (food_id IS NOT NULL OR recipe_id IS NOT NULL),
  CONSTRAINT uq_meal_items_meal_client UNIQUE (meal_id, client_id)
);
CREATE INDEX idx_meal_items_food ON meal_items (food_id);
CREATE INDEX idx_meal_items_recipe ON meal_items (recipe_id);
CREATE INDEX idx_meal_items_cursor ON meal_items (server_updated_at);

CREATE TRIGGER meal_items_set_server_updated_at BEFORE INSERT OR UPDATE ON meal_items
  FOR EACH ROW EXECUTE FUNCTION pesa_touch_server_updated_at();

-- +goose Down
DROP TABLE meal_items;
