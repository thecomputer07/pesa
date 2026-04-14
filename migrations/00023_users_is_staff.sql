-- +goose Up
-- LEGAL-10: add is_staff flag to users (used by RequireStaff middleware).
ALTER TABLE users ADD COLUMN is_staff BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users DROP COLUMN is_staff;
