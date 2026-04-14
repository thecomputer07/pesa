-- +goose Up
CREATE TABLE recipe_items (
  id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  recipe_id         BIGINT NOT NULL,
  client_id         CHAR(36) NOT NULL,
  food_id           BIGINT NULL,
  name_snapshot     VARCHAR(255) NOT NULL,
  grams             NUMERIC(7,2) NOT NULL,
  snapshot_json     JSONB NOT NULL,
  updated_at_ms     BIGINT NOT NULL,
  deleted_at_ms     BIGINT NULL,
  server_updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_recipe_items_recipe FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,
  CONSTRAINT fk_recipe_items_food   FOREIGN KEY (food_id)   REFERENCES foods(id),
  CONSTRAINT uq_recipe_items_recipe_client UNIQUE (recipe_id, client_id)
);
CREATE INDEX idx_recipe_items_food ON recipe_items (food_id);
CREATE INDEX idx_recipe_items_cursor ON recipe_items (server_updated_at);

CREATE TRIGGER recipe_items_set_server_updated_at BEFORE INSERT OR UPDATE ON recipe_items
  FOR EACH ROW EXECUTE FUNCTION pesa_touch_server_updated_at();

-- +goose Down
DROP TABLE recipe_items;
