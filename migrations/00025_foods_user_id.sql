-- +goose Up
-- Consente agli utenti di aggiungere propri alimenti personalizzati.
-- user_id IS NULL  → alimento condiviso (CREA, OFF, fixture)
-- user_id IS NOT NULL → alimento privato dell'utente (fonte "manual")
ALTER TABLE foods ADD COLUMN user_id BIGINT NULL
  REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_foods_user_id ON foods (user_id) WHERE user_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_foods_user_id;
ALTER TABLE foods DROP COLUMN user_id;
