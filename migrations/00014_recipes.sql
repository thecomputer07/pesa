-- +goose Up
CREATE TABLE recipes (
  id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id           BIGINT NOT NULL,
  client_id         CHAR(36) NOT NULL,
  name              VARCHAR(255) NOT NULL,
  servings          NUMERIC(5,2) NOT NULL DEFAULT 1,
  total_grams       NUMERIC(7,2) NULL,
  notes             TEXT NULL,
  updated_at_ms     BIGINT NOT NULL,
  deleted_at_ms     BIGINT NULL,
  server_updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_recipes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT uq_recipes_user_client UNIQUE (user_id, client_id)
);
CREATE INDEX idx_recipes_user_cursor ON recipes (user_id, server_updated_at);

CREATE TRIGGER recipes_set_server_updated_at BEFORE INSERT OR UPDATE ON recipes
  FOR EACH ROW EXECUTE FUNCTION pesa_touch_server_updated_at();

ALTER TABLE meal_items
  ADD CONSTRAINT fk_meal_items_recipe FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE meal_items DROP CONSTRAINT IF EXISTS fk_meal_items_recipe;
DROP TABLE recipes;
