-- +goose Up
CREATE TABLE meal_templates (
  id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id           BIGINT NOT NULL,
  client_id         CHAR(36) NOT NULL,
  name              VARCHAR(255) NOT NULL,
  payload_json      JSONB NOT NULL,
  updated_at_ms     BIGINT NOT NULL,
  deleted_at_ms     BIGINT NULL,
  server_updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_meal_templates_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT uq_meal_templates_user_client UNIQUE (user_id, client_id)
);
CREATE INDEX idx_meal_templates_user_cursor ON meal_templates (user_id, server_updated_at);

CREATE TRIGGER meal_templates_set_server_updated_at BEFORE INSERT OR UPDATE ON meal_templates
  FOR EACH ROW EXECUTE FUNCTION pesa_touch_server_updated_at();

-- +goose Down
DROP TABLE meal_templates;
