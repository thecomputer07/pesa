-- +goose Up
CREATE TABLE food_images (
  id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  food_id      BIGINT NOT NULL,
  url          VARCHAR(512) NOT NULL,
  attribution  VARCHAR(255) NULL,
  license      VARCHAR(64) NULL,
  is_external  BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_food_images_food FOREIGN KEY (food_id) REFERENCES foods(id) ON DELETE CASCADE
);
CREATE INDEX idx_food_images_food ON food_images (food_id);

-- +goose Down
DROP TABLE food_images;
