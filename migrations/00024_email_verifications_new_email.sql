-- +goose Up
-- PRIV-05: optional new_email column for email-change swaps.
ALTER TABLE email_verifications ADD COLUMN new_email VARCHAR(254) NULL;

-- +goose Down
ALTER TABLE email_verifications DROP COLUMN new_email;
