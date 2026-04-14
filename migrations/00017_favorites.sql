-- +goose Up
CREATE TABLE favorites (
  id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id           BIGINT NOT NULL,
  client_id         CHAR(36) NOT NULL,
  food_id           BIGINT NOT NULL,
  updated_at_ms     BIGINT NOT NULL,
  deleted_at_ms     BIGINT NULL,
  server_updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_favorites_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_favorites_food FOREIGN KEY (food_id) REFERENCES foods(id) ON DELETE CASCADE,
  CONSTRAINT uq_favorites_user_client UNIQUE (user_id, client_id),
  CONSTRAINT uq_favorites_user_food   UNIQUE (user_id, food_id)
);
CREATE INDEX idx_favorites_user_cursor ON favorites (user_id, server_updated_at);

CREATE TRIGGER favorites_set_server_updated_at BEFORE INSERT OR UPDATE ON favorites
  FOR EACH ROW EXECUTE FUNCTION pesa_touch_server_updated_at();

-- +goose Down
DROP TABLE favorites;
