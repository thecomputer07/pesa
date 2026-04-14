-- +goose Up
CREATE TABLE meals (
  id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id           BIGINT NOT NULL,
  client_id         CHAR(36) NOT NULL,
  occurred_at       TIMESTAMPTZ NOT NULL,
  slot              TEXT NOT NULL CHECK (slot IN ('colazione','pranzo','cena','spuntino','altro')),
  notes             VARCHAR(500) NULL,
  updated_at_ms     BIGINT NOT NULL,
  deleted_at_ms     BIGINT NULL,
  server_updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_meals_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT uq_meals_user_client UNIQUE (user_id, client_id)
);
CREATE INDEX idx_meals_user_occurred ON meals (user_id, occurred_at);
CREATE INDEX idx_meals_user_cursor ON meals (user_id, server_updated_at);

CREATE TRIGGER meals_set_server_updated_at BEFORE INSERT OR UPDATE ON meals
  FOR EACH ROW EXECUTE FUNCTION pesa_touch_server_updated_at();

-- +goose Down
DROP TABLE meals;
